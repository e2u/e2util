package e2aws

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSession(t *testing.T) {
	t.Run("loads config with region and credentials", func(t *testing.T) {
		cfg := NewSession("ap-northeast-1",
			config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("AKID", "SECRET", "TOKEN")),
		)
		assert.Equal(t, "ap-northeast-1", cfg.Region)
		require.NotNil(t, cfg.Credentials)
		creds, err := cfg.Credentials.Retrieve(t.Context())
		require.NoError(t, err)
		assert.Equal(t, "AKID", creds.AccessKeyID)
		assert.Equal(t, "SECRET", creds.SecretAccessKey)
		assert.Equal(t, "TOKEN", creds.SessionToken)
	})

	t.Run("returns empty config when load option fails", func(t *testing.T) {
		cfg := NewSession("us-east-1", func(*config.LoadOptions) error {
			return errors.New("load failed")
		})
		assert.Equal(t, aws.Config{}, cfg)
	})
}

func TestGetHostName(t *testing.T) {
	want, err := os.Hostname()
	require.NoError(t, err)

	got, err := GetHostName()
	require.NoError(t, err)
	assert.Equal(t, want, got)
}

func TestMustGetHostName(t *testing.T) {
	orig := hostName
	t.Cleanup(func() { hostName = orig })

	t.Run("returns cached value", func(t *testing.T) {
		hostName = "cached-host"
		assert.Equal(t, "cached-host", MustGetHostName())
	})

	t.Run("resolves and caches hostname", func(t *testing.T) {
		hostName = ""
		want, err := os.Hostname()
		require.NoError(t, err)
		assert.Equal(t, want, MustGetHostName())
		assert.Equal(t, want, hostName)
		assert.Equal(t, want, MustGetHostName())
	})
}

func TestGetIP(t *testing.T) {
	ip, err := GetIP()
	if err != nil {
		assert.Equal(t, "are you connected to the network?", err.Error())
		assert.Empty(t, ip)
		return
	}
	assert.NotEmpty(t, ip)
}

func TestMustGetIP(t *testing.T) {
	orig := hostIP
	t.Cleanup(func() { hostIP = orig })

	t.Run("returns cached value", func(t *testing.T) {
		hostIP = "10.0.0.9"
		assert.Equal(t, "10.0.0.9", MustGetIP())
	})

	t.Run("resolves and caches ip when available", func(t *testing.T) {
		hostIP = ""
		ip, err := GetIP()
		if err != nil {
			assert.Empty(t, MustGetIP())
			return
		}
		assert.Equal(t, ip, MustGetIP())
		assert.Equal(t, ip, hostIP)
		assert.Equal(t, ip, MustGetIP())
	})
}

type hostRewriteTransport struct {
	target *url.URL
	base   http.RoundTripper
}

func (h hostRewriteTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	clone := req.Clone(req.Context())
	clone.URL.Scheme = h.target.Scheme
	clone.URL.Host = h.target.Host
	clone.Host = h.target.Host
	return h.base.RoundTrip(clone)
}

func testAWSConfig(serverURL string, httpClient aws.HTTPClient) aws.Config {
	return aws.Config{
		Region:                     "us-east-1",
		Credentials:                credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
		BaseEndpoint:               aws.String(serverURL),
		HTTPClient:                 httpClient,
		Retryer:                    func() aws.Retryer { return aws.NopRetryer{} },
		RequestChecksumCalculation: aws.RequestChecksumCalculationWhenRequired,
		ResponseChecksumValidation: aws.ResponseChecksumValidationWhenRequired,
	}
}

func TestUploadToS3(t *testing.T) {
	t.Run("missing local file", func(t *testing.T) {
		err := UploadToS3(aws.Config{Region: "us-east-1"}, filepath.Join(t.TempDir(), "missing.txt"), "bucket", "key")
		require.Error(t, err)
	})

	t.Run("uploads file contents", func(t *testing.T) {
		var gotMethod, gotPath string
		var gotBody []byte
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotMethod = r.Method
			gotPath = r.URL.Path
			b, _ := io.ReadAll(r.Body)
			gotBody = b
			w.Header().Set("ETag", `"abc"`)
			w.WriteHeader(http.StatusOK)
		}))
		t.Cleanup(srv.Close)

		dir := t.TempDir()
		local := filepath.Join(dir, "hello.txt")
		require.NoError(t, os.WriteFile(local, []byte("hello s3"), 0o600))

		target, err := url.Parse(srv.URL)
		require.NoError(t, err)
		client := &http.Client{Transport: hostRewriteTransport{target: target, base: srv.Client().Transport}}
		cfg := testAWSConfig(srv.URL, client)

		err = UploadToS3(cfg, local, "my-bucket", "path/hello.txt")
		require.NoError(t, err)
		assert.Equal(t, http.MethodPut, gotMethod)
		assert.Contains(t, gotPath, "hello.txt")
		assert.Equal(t, []byte("hello s3"), gotBody)
	})
}
