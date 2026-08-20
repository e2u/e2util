package e2s3

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3 结构
type S3 struct {
	client *s3.Client
}

// New 生成一个新的实例
func New(cfg aws.Config, optFns ...func(*s3.Options)) *S3 {
	return &S3{
		client: s3.NewFromConfig(cfg, optFns...),
	}
}

func (s *S3) instance() *s3.Client {
	return s.client
}

func (s *S3) presign() *s3.PresignClient {
	return s3.NewPresignClient(s.client)
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
func (s *S3) ListBucketFiles(bucketName, prefix string, fn func(objs []types.Object, lastPage bool)) error {
	p := s3.NewListObjectsV2Paginator(s.client, &s3.ListObjectsV2Input{
		Bucket: aws.String(bucketName),
		Prefix: aws.String(prefix),
	})
	for p.HasMorePages() {
		page, err := p.NextPage(context.Background())
		if err != nil {
			return err
		}
		fn(page.Contents, !p.HasMorePages())
	}
	return nil
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

	_, err := s.instance().PutObject(context.Background(), si)
	return err
}

// PreSignedGetObjectURL 生成获取对象地址的预签名地址
func (s *S3) PreSignedGetObjectURL(bucketName, key string, expires time.Duration) (string, error) {
	out, err := s.presign().PresignGetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    s.fixKey(key),
	}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

// PreSignedPutObjectURL 生成上传对象地址的预签名地址
func (s *S3) PreSignedPutObjectURL(bucketName, key string, expires time.Duration, opts ...*s3.PutObjectInput) (string, error) {
	pi := &s3.PutObjectInput{}
	if len(opts) > 0 {
		pi = opts[0]
	}
	pi.Bucket = aws.String(bucketName)
	pi.Key = s.fixKey(key)
	out, err := s.presign().PresignPutObject(context.Background(), pi, s3.WithPresignExpires(expires))
	if err != nil {
		return "", err
	}
	return out.URL, nil
}

// GetObject 读取一个对象
func (s *S3) GetObject(bucketName, key string, opts ...*s3.GetObjectInput) ([]byte, error) {
	pi := &s3.GetObjectInput{}
	if len(opts) > 0 {
		pi = opts[0]
	}
	pi.Bucket = aws.String(bucketName)
	pi.Key = s.fixKey(key)

	out, err := s.instance().GetObject(context.Background(), pi)
	if err != nil {
		return nil, err
	}
	defer func() { _ = out.Body.Close() }()
	b, err := io.ReadAll(out.Body)
	return b, err
}

// UploadWithFilePath 上传本地文件到 s3 上,无法设置更多属性，如需定制，请用 Upload 方法
func (s *S3) UploadWithFilePath(localFile, bucket, key string) (string, error) {
	file, err := os.Open(filepath.Clean(localFile))
	if err != nil {
		return "", err
	}
	defer func() { _ = file.Close() }()
	svc := manager.NewUploader(s.client)
	out, err := svc.Upload(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    s.fixKey(key),
		Body:   file,
	})
	if out != nil {
		return out.Location, err
	}
	return "", err
}

// Upload 上传内容到 s3
func (s *S3) Upload(bucket, key string, input *s3.PutObjectInput) (string, error) {
	input.Bucket = aws.String(bucket)
	input.Key = s.fixKey(key)
	svc := manager.NewUploader(s.client)
	out, err := svc.Upload(context.Background(), input)
	if out != nil {
		return out.Location, err
	}
	return "", err
}

func (s *S3) fixKey(key string) *string {
	for strings.HasPrefix(key, "/") && len(key) != 0 {
		key = key[1:]
	}
	return aws.String(key)
}

// DeleteObject 刪除一個對象
func (s *S3) DeleteObject(bucket, key string) error {
	_, err := s.instance().DeleteObject(context.Background(), &s3.DeleteObjectInput{
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
	_, err = s.instance().CopyObject(context.Background(), &s3.CopyObjectInput{
		Bucket:     aws.String(bucket),
		Key:        aws.String(key),
		CopySource: aws.String(url.PathEscape(srcObject)),
	})
	return err
}
