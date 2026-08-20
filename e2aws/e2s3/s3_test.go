package e2s3

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testConfig() aws.Config {
	return aws.Config{
		Region:                     "us-east-1",
		Credentials:                credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}
}

func newTestS3(t *testing.T, handler http.HandlerFunc) *S3 {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(testConfig(), func(o *s3.Options) {
		o.BaseEndpoint = aws.String(srv.URL)
		o.HTTPClient = srv.Client()
		o.UsePathStyle = true
		o.Retryer = aws.NopRetryer{}
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.ResponseChecksumValidation = aws.ResponseChecksumValidationWhenRequired
	})
}

func s3Error(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, `<Error><Code>`+code+`</Code><Message>`+message+`</Message></Error>`)
}

func TestNew(t *testing.T) {
	s := New(testConfig(), func(o *s3.Options) {
		o.Region = "ap-southeast-1"
		o.UsePathStyle = true
	})
	require.NotNil(t, s)
	assert.NotNil(t, s.client)
	assert.Equal(t, "ap-southeast-1", s.client.Options().Region)
	assert.True(t, s.client.Options().UsePathStyle)
	assert.Equal(t, s.client, s.instance())
	assert.NotNil(t, s.presign())
}

func TestParseS3Path(t *testing.T) {
	s := &S3{}

	tests := []struct {
		name       string
		in         string
		wantBucket string
		wantKey    string
		wantErr    string
	}{
		{name: "bucket and nested key", in: "s3://my-bucket/path/to/file.txt", wantBucket: "my-bucket", wantKey: "path/to/file.txt"},
		{name: "bucket only", in: "s3://my-bucket", wantBucket: "my-bucket", wantKey: ""},
		{name: "bucket with slash only", in: "s3://my-bucket/", wantBucket: "my-bucket", wantKey: "/"},
		{name: "missing scheme", in: "https://example.com/file", wantErr: "illegal parameter"},
		{name: "invalid url", in: "s3://%", wantErr: ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bucket, key, err := s.ParseS3Path(tt.in)
			if tt.wantErr != "" || strings.Contains(tt.in, "%") && tt.name == "invalid url" {
				require.Error(t, err)
				if tt.wantErr != "" {
					assert.Contains(t, err.Error(), tt.wantErr)
				}
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantBucket, bucket)
			assert.Equal(t, tt.wantKey, key)
		})
	}
}

func TestFixKey(t *testing.T) {
	s := &S3{}
	assert.Equal(t, "a/b", aws.ToString(s.fixKey("/a/b")))
	assert.Equal(t, "a/b", aws.ToString(s.fixKey("///a/b")))
	assert.Equal(t, "a/b", aws.ToString(s.fixKey("a/b")))
	assert.Equal(t, "", aws.ToString(s.fixKey("/")))
	assert.Equal(t, "", aws.ToString(s.fixKey("")))
}

func TestListBucketFiles(t *testing.T) {
	t.Run("paginates objects", func(t *testing.T) {
		var calls atomic.Int32
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, http.MethodGet, r.Method)
			assert.Equal(t, "2", r.URL.Query().Get("list-type"))
			assert.Equal(t, "pre/", r.URL.Query().Get("prefix"))
			n := calls.Add(1)
			w.Header().Set("Content-Type", "application/xml")
			if n == 1 {
				_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult>
  <Name>my-bucket</Name>
  <Prefix>pre/</Prefix>
  <IsTruncated>true</IsTruncated>
  <NextContinuationToken>token-1</NextContinuationToken>
  <Contents><Key>pre/a.txt</Key><Size>1</Size></Contents>
</ListBucketResult>`)
				return
			}
			assert.Equal(t, "token-1", r.URL.Query().Get("continuation-token"))
			_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>
<ListBucketResult>
  <Name>my-bucket</Name>
  <Prefix>pre/</Prefix>
  <IsTruncated>false</IsTruncated>
  <Contents><Key>pre/b.txt</Key><Size>2</Size></Contents>
</ListBucketResult>`)
		})

		var keys []string
		var lastFlags []bool
		err := s.ListBucketFiles("my-bucket", "pre/", func(objs []types.Object, lastPage bool) {
			for _, obj := range objs {
				keys = append(keys, aws.ToString(obj.Key))
			}
			lastFlags = append(lastFlags, lastPage)
		})
		require.NoError(t, err)
		assert.Equal(t, []string{"pre/a.txt", "pre/b.txt"}, keys)
		assert.Equal(t, []bool{false, true}, lastFlags)
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			s3Error(w, http.StatusForbidden, "AccessDenied", "denied")
		})
		err := s.ListBucketFiles("my-bucket", "pre/", func([]types.Object, bool) {})
		require.Error(t, err)
	})
}

func TestPutContentObject(t *testing.T) {
	t.Run("writes body and strips leading slash", func(t *testing.T) {
		var gotPath, gotType string
		var gotBody []byte
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotType = r.Header.Get("Content-Type")
			gotBody, _ = io.ReadAll(r.Body)
			w.Header().Set("ETag", `"etag"`)
			w.WriteHeader(http.StatusOK)
		})
		err := s.PutContentObject("my-bucket", "/dir/file.txt", []byte("payload"), &s3.PutObjectInput{
			ContentType: aws.String("text/plain"),
		})
		require.NoError(t, err)
		assert.Equal(t, "/my-bucket/dir/file.txt", gotPath)
		assert.Equal(t, "text/plain", gotType)
		assert.Equal(t, []byte("payload"), gotBody)
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			s3Error(w, http.StatusForbidden, "AccessDenied", "denied")
		})
		err := s.PutContentObject("my-bucket", "file.txt", []byte("x"))
		require.Error(t, err)
	})
}

type errCreds struct{}

func (errCreds) Retrieve(ctx context.Context) (aws.Credentials, error) {
	return aws.Credentials{}, errors.New("no credentials")
}

func TestPreSignedGetObjectURL(t *testing.T) {
	t.Run("signs get url", func(t *testing.T) {
		s := New(testConfig())
		u, err := s.PreSignedGetObjectURL("my-bucket", "/path/file.txt", time.Minute)
		require.NoError(t, err)
		assert.Contains(t, u, "my-bucket")
		assert.Contains(t, u, "path/file.txt")
		assert.Contains(t, u, "X-Amz-Signature=")
		assert.Contains(t, u, "X-Amz-Expires=")
	})

	t.Run("credential error", func(t *testing.T) {
		s := New(aws.Config{Region: "us-east-1", Credentials: errCreds{}})
		u, err := s.PreSignedGetObjectURL("my-bucket", "file.txt", time.Minute)
		require.Error(t, err)
		assert.Empty(t, u)
	})
}

func TestPreSignedPutObjectURL(t *testing.T) {
	t.Run("signs put url", func(t *testing.T) {
		s := New(testConfig())
		u, err := s.PreSignedPutObjectURL("my-bucket", "/path/file.txt", 2*time.Minute, &s3.PutObjectInput{
			ContentType: aws.String("application/json"),
		})
		require.NoError(t, err)
		assert.Contains(t, u, "my-bucket")
		assert.Contains(t, u, "path/file.txt")
		assert.Contains(t, u, "X-Amz-Signature=")
	})

	t.Run("credential error", func(t *testing.T) {
		s := New(aws.Config{Region: "us-east-1", Credentials: errCreds{}})
		u, err := s.PreSignedPutObjectURL("my-bucket", "file.txt", time.Minute)
		require.Error(t, err)
		assert.Empty(t, u)
	})
}

func TestGetObject(t *testing.T) {
	t.Run("reads object body", func(t *testing.T) {
		var gotPath string
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			w.Header().Set("Content-Type", "text/plain")
			_, _ = io.WriteString(w, "hello object")
		})
		b, err := s.GetObject("my-bucket", "/dir/file.txt", &s3.GetObjectInput{})
		require.NoError(t, err)
		assert.Equal(t, []byte("hello object"), b)
		assert.Equal(t, "/my-bucket/dir/file.txt", gotPath)
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			s3Error(w, http.StatusNotFound, "NoSuchKey", "missing")
		})
		_, err := s.GetObject("my-bucket", "missing.txt")
		require.Error(t, err)
	})
}

func TestUploadWithFilePath(t *testing.T) {
	t.Run("missing local file", func(t *testing.T) {
		s := New(testConfig())
		_, err := s.UploadWithFilePath(filepath.Join(t.TempDir(), "nope.txt"), "bucket", "key")
		require.Error(t, err)
	})

	t.Run("uploads file", func(t *testing.T) {
		var gotPath string
		var gotBody []byte
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotBody, _ = io.ReadAll(r.Body)
			w.Header().Set("ETag", `"etag"`)
			w.WriteHeader(http.StatusOK)
		})
		local := filepath.Join(t.TempDir(), "local.txt")
		require.NoError(t, os.WriteFile(local, []byte("from disk"), 0o600))
		loc, err := s.UploadWithFilePath(local, "my-bucket", "/dir/local.txt")
		require.NoError(t, err)
		assert.NotEmpty(t, loc)
		assert.Equal(t, "/my-bucket/dir/local.txt", gotPath)
		assert.Equal(t, []byte("from disk"), gotBody)
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			s3Error(w, http.StatusForbidden, "AccessDenied", "denied")
		})
		local := filepath.Join(t.TempDir(), "local.txt")
		require.NoError(t, os.WriteFile(local, []byte("from disk"), 0o600))
		loc, err := s.UploadWithFilePath(local, "my-bucket", "dir/local.txt")
		require.Error(t, err)
		assert.Empty(t, loc)
	})
}

func TestUpload(t *testing.T) {
	t.Run("uploads reader", func(t *testing.T) {
		var gotPath string
		var gotBody []byte
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotBody, _ = io.ReadAll(r.Body)
			w.Header().Set("ETag", `"etag"`)
			w.WriteHeader(http.StatusOK)
		})
		loc, err := s.Upload("my-bucket", "/dir/mem.txt", &s3.PutObjectInput{
			Body: bytes.NewReader([]byte("memory")),
		})
		require.NoError(t, err)
		assert.NotEmpty(t, loc)
		assert.Equal(t, "/my-bucket/dir/mem.txt", gotPath)
		assert.Equal(t, []byte("memory"), gotBody)
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			s3Error(w, http.StatusForbidden, "AccessDenied", "denied")
		})
		_, err := s.Upload("my-bucket", "file.txt", &s3.PutObjectInput{Body: bytes.NewReader([]byte("x"))})
		require.Error(t, err)
	})
}

func TestDeleteObject(t *testing.T) {
	t.Run("deletes object", func(t *testing.T) {
		var gotMethod, gotPath string
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			w.WriteHeader(http.StatusNoContent)
		})
		err := s.DeleteObject("my-bucket", "dir/file.txt")
		require.NoError(t, err)
		assert.Equal(t, http.MethodDelete, gotMethod)
		assert.Equal(t, "/my-bucket/dir/file.txt", gotPath)
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			s3Error(w, http.StatusForbidden, "AccessDenied", "denied")
		})
		err := s.DeleteObject("my-bucket", "file.txt")
		require.Error(t, err)
	})
}

func TestCopyObject(t *testing.T) {
	t.Run("invalid target path", func(t *testing.T) {
		s := &S3{}
		err := s.CopyObject("s3://src/a", "not-s3")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "illegal parameter")
	})

	t.Run("strips s3 scheme from source", func(t *testing.T) {
		var gotPath, gotSource string
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			gotPath = r.URL.Path
			gotSource = r.Header.Get("x-amz-copy-source")
			w.Header().Set("Content-Type", "application/xml")
			_, _ = io.WriteString(w, `<CopyObjectResult><ETag>"etag"</ETag><LastModified>2020-01-01T00:00:00.000Z</LastModified></CopyObjectResult>`)
		})
		err := s.CopyObject("s3://src-bucket/from/key.txt", "s3://dst-bucket/to/key.txt")
		require.NoError(t, err)
		assert.Equal(t, "/dst-bucket/to/key.txt", gotPath)
		assert.Equal(t, "src-bucket%2Ffrom%2Fkey.txt", gotSource)
	})

	t.Run("keeps source without s3 scheme", func(t *testing.T) {
		var gotSource string
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			gotSource = r.Header.Get("x-amz-copy-source")
			w.Header().Set("Content-Type", "application/xml")
			_, _ = io.WriteString(w, `<CopyObjectResult><ETag>"etag"</ETag></CopyObjectResult>`)
		})
		err := s.CopyObject("src-bucket/from/key.txt", "s3://dst-bucket/to/key.txt")
		require.NoError(t, err)
		assert.Equal(t, "src-bucket%2Ffrom%2Fkey.txt", gotSource)
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestS3(t, func(w http.ResponseWriter, r *http.Request) {
			s3Error(w, http.StatusNotFound, "NoSuchKey", "missing")
		})
		err := s.CopyObject("s3://src/a", "s3://dst/b")
		require.Error(t, err)
	})
}
