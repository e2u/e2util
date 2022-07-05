package e2aws

import (
	"errors"
	"net"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/sirupsen/logrus"
)

var (
	hostIP   string
	hostName string
)

// NewSession 返回一个 aws session
// 默认不输出日志 ，如需日志输出，可以在 cfgs 参数中传入
// 	cfgs := aws.NewConfig().
//		WithLogLevel(aws.LogDebugWithHTTPBody).
//		WithLogger(aws.LoggerFunc(func(args ...interface{}) {
//			fmt.Fprintln(os.Stdout, args...)
//		}))
func NewSession(region string, cfgs ...*aws.Config) *session.Session {
	cfg := aws.NewConfig().
		WithRegion(region).
		WithCredentialsChainVerboseErrors(true).
		WithLogLevel(aws.LogOff)

	sess, err := session.NewSession(cfg)
	if err != nil {
		logrus.Errorf("aws new session error=%v", err)
		return nil
	}
	return sess

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

// GetIP 获取当前运行 ec2 实例的 ip
func GetIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("are you connected to the network?")
}

// UploadToS3 不推荐再使用
func UploadToS3(sess *session.Session, localfile, bucket, s3path string) error {
	file, err := os.Open(localfile)
	if err != nil {
		return err
	}
	svc := s3manager.NewUploader(sess)
	_, err = svc.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(s3path),
		Body:   file,
	})
	return err
}
