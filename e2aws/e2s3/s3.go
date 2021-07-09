package e2s3

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

// S3 结构
type S3 struct {
	sess *session.Session
	cfgs *aws.Config
}

// New 生成一个新的实例
func New(sess *session.Session, cfgs ...*aws.Config) *S3 {
	sess = sess.Copy(cfgs...)
	return &S3{
		sess: sess,
	}
}

func (s *S3) instance() *s3.S3 {
	_ = s.cfgs
	return s3.New(s.sess)
}

// ParseS3Path 解析一个完整的 s3 路径, s3://bucket/key/subkey/file
// 返回: bucket name,key ,error
func (s *S3) ParseS3Path(s3path string) (string, string, error) {
	if !strings.HasPrefix(s3path, "s3://") {
		return "", "", fmt.Errorf(" illegal parameter %v", s3path)
	}
	u, err := url.Parse(s3path)
	if err != nil {
		return "", "", err
	}
	path := u.Path
	if strings.HasPrefix(path, "/") && len(path) > 1 {
		path = path[1:]
	}
	return u.Host, path, nil
}

// ListBucketFiles 列出指定桶下的符合条件的文件
func (s *S3) ListBucketFiles(bucketName, prefix string, fn func(objs []*s3.Object, lastPage bool)) error {
	return s.instance().ListObjectsPages(&s3.ListObjectsInput{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	}, func(output *s3.ListObjectsOutput, lastPage bool) bool {
		fn(output.Contents, lastPage)
		return lastPage
	})
}

// PutContentObject 写数据到 s3 中
func (s *S3) PutContentObject(bucketName string, key string, content []byte, opts ...*s3.PutObjectInput) error {
	si := &s3.PutObjectInput{}
	if len(opts) > 0 {
		si = opts[0]
	}
	si.Bucket = aws.String(bucketName)
	si.Key = s.fixKey(key)
	si.Body = bytes.NewReader(content)

	_, err := s.instance().PutObject(si)
	return err
}

// PreSignedGetObjectURL 生成获取对象地址的预签名地址
func (s *S3) PreSignedGetObjectURL(bucketName, key string, expires time.Duration) (string, error) {
	req, _ := s.instance().GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    s.fixKey(key),
	})
	return req.Presign(expires)
}

// PreSignedPutObjectURL 生成上传对象地址的预签名地址
func (s *S3) PreSignedPutObjectURL(bucketName, key string, expires time.Duration, opts ...*s3.PutObjectInput) (string, error) {
	pi := &s3.PutObjectInput{}
	if len(opts) > 0 {
		pi = opts[0]
	}
	pi.Bucket = aws.String(bucketName)
	pi.Key = s.fixKey(key)
	req, _ := s.instance().PutObjectRequest(pi)
	return req.Presign(expires)
}

// GetObject 读取一个对象
func (s *S3) GetObject(bucketName, key string, opts ...*s3.GetObjectInput) ([]byte, error) {
	pi := &s3.GetObjectInput{}
	if len(opts) > 0 {
		pi = opts[0]
	}
	pi.Bucket = aws.String(bucketName)
	pi.Key = s.fixKey(key)

	out, err := s.instance().GetObject(pi)
	if err != nil {
		return nil, err
	}
	defer func() { _ = out.Body.Close() }()
	b, err := ioutil.ReadAll(out.Body)
	return b, err
}

// UploadWithFilePath 上传本地文件到 s3 上,无法设置更多属性，如需定制，请用 Upload 方法
func (s *S3) UploadWithFilePath(localFile, bucket, key string) (string, error) {
	file, err := os.Open(localFile)
	if err != nil {
		return "", err
	}
	svc := s3manager.NewUploader(s.sess)
	out, err := svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    s.fixKey(key),
		Body:   file,
	})
	return func() (string, error) {
		if out != nil {
			return out.Location, err
		}
		return "", err
	}()
}

// Upload 上传内容到 s3
func (s *S3) Upload(bucket, key string, input *s3manager.UploadInput) (string, error) {
	input.Bucket = aws.String(bucket)
	input.Key = s.fixKey(key)
	svc := s3manager.NewUploader(s.sess)
	out, err := svc.Upload(input)
	return func() (string, error) {
		if out != nil {
			return out.Location, err
		}
		return "", err
	}()
}

func (s *S3) fixKey(key string) *string {
	for {
		if !strings.HasPrefix(key, "/") || len(key) == 0 {
			break
		}
		key = key[1:]
	}
	return aws.String(key)
}

// DeleteObject 刪除一個對象
func (s *S3) DeleteObject(bucket, key string) error {
	_, err := s.instance().DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	return err
}

// CopyObject 複製一個對象
// srcObject s3://bucket/key
// target s3://bucket/key
func (s *S3) CopyObject(srcObject, targetObject string) error {
	bucket, key, err := s.ParseS3Path(targetObject)
	if err != nil {
		return err
	}
	if strings.HasPrefix(srcObject, "s3://") {
		srcObject = strings.Replace(srcObject, "s3://", "", 1)
	}
	_, err = s.instance().CopyObject(&s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(key),
		CopySource: aws.String(url.PathEscape(srcObject)),
	})
	return err
}
