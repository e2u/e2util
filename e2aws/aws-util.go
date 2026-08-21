package e2aws

import (
	"context"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/e2u/e2util/e2os"
	"github.com/sirupsen/logrus"
)

var (
	hostIP   string
	hostName string
)

// NewSession 返回一个 aws.Config（AWS SDK v2）。
// 默认不输出日志 ，如需日志输出，可以在 optFns 参数中传入
//
//	cfg := e2aws.NewSession("us-east-1",
//		config.WithClientLogMode(aws.LogRequestWithBody|aws.LogResponseWithBody),
//	)
func NewSession(region string, optFns ...func(*config.LoadOptions) error) aws.Config {
	opts := make([]func(*config.LoadOptions) error, 0, 1+len(optFns))
	opts = append(opts, config.WithRegion(region))
	opts = append(opts, optFns...)

	cfg, err := config.LoadDefaultConfig(context.Background(), opts...)
	if err != nil {
		logrus.Errorf("aws new session error=%v", err)
		return aws.Config{}
	}
	return cfg
}

func GetHostName() (string, error) {
	return os.Hostname()
}

func MustGetHostName() string {
	if len(hostName) > 0 {
		return hostName
	}
	if h, err := GetHostName(); err == nil {
		hostName = h
		return h
	}
	return ""
}

// MustGetIP 获取当前运行 ec2 实例的 ip
func MustGetIP() string {
	if len(hostIP) > 0 {
		return hostIP
	}
	if i, err := GetIP(); err == nil {
		hostIP = i
		return i
	}
	return ""
}

// GetIP 获取当前运行实例的 ipv4，实现见 e2os.ExternalIP
func GetIP() (string, error) {
	return e2os.ExternalIP()
}

// UploadToS3 不推荐再使用
func UploadToS3(cfg aws.Config, localfile, bucket, s3path string) error {
	f, err := os.Open(filepath.Clean(localfile))
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	svc := manager.NewUploader(s3.NewFromConfig(cfg))
	_, err = svc.Upload(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3path),
		Body:   f,
	})
	return err
}
