package e2s3v2

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3 struct {
	client *s3.Client
}

func New(cfg aws.Config, optFns ...func(*s3.Options)) *S3 {
	return &S3{
		client: s3.NewFromConfig(cfg, optFns...),
	}
}

// func (s *S3) instance() *s3.Client {
//	return s.client
//}

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
