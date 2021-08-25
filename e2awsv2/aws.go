package e2awsv2

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
)

func init() {
	config.LoadDefaultConfig(context.TODO())
}
