package e2sqs

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"github.com/e2u/e2util/e2crypto"
	"github.com/e2u/e2util/e2exec"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
)

type SQS struct {
	urlCache sync.Map
	client   *sqs.Client
}

func New(cfg aws.Config, optFns ...func(*sqs.Options)) *SQS {
	return &SQS{
		client: sqs.NewFromConfig(cfg, optFns...),
	}
}

func (s *SQS) instance() *sqs.Client {
	return s.client
}

func (s *SQS) GetURL(queueName string) (*string, error) {
	if v, ok := s.urlCache.Load(queueName); ok {
		return v.(*string), nil
	}
	out, err := s.instance().GetQueueUrl(context.Background(), &sqs.GetQueueUrlInput{QueueName: aws.String(queueName)})
	if err != nil {
		return nil, err
	}
	s.urlCache.Store(queueName, out.QueueUrl)
	return out.QueueUrl, nil
}

func (s *SQS) MustGetURL(queueName string) *string {
	u, _ := s.GetURL(queueName)
	return u
}

// SendMessage 发送单条消息到队列
func (s *SQS) SendMessage(queueName string, message string) error {
	_, err := s.instance().SendMessage(context.Background(), &sqs.SendMessageInput{
		QueueUrl:    s.MustGetURL(queueName),
		MessageBody: aws.String(message),
	})
	return err
}

// BatchSendMessages 批量发送消息到队列
func (s *SQS) BatchSendMessages(queueName string, messages []string) (int, error) {
	entries := make([]types.SendMessageBatchRequestEntry, 0, len(messages))
	for _, message := range messages {
		id := e2exec.Must(e2crypto.RandomString(16))
		msg := message
		entries = append(entries, types.SendMessageBatchRequestEntry{MessageBody: aws.String(msg), Id: aws.String(id)})
	}
	out, err := s.instance().SendMessageBatch(context.Background(), &sqs.SendMessageBatchInput{
		QueueUrl: s.MustGetURL(queueName),
		Entries:  entries,
	})
	if out == nil {
		return 0, err
	}
	return len(messages) - len(out.Failed), err
}

// ReceiveMessage 接收队列的消息
func (s *SQS) ReceiveMessage(queueName string, maxNumber int64) ([]types.Message, error) {
	out, err := s.instance().ReceiveMessage(context.Background(), &sqs.ReceiveMessageInput{
		QueueUrl:            s.MustGetURL(queueName),
		MaxNumberOfMessages: int32(maxNumber),
	})
	if out == nil {
		return nil, err
	}
	return out.Messages, err
}

// DeleteMessage 删除单条消息
func (s *SQS) DeleteMessage(queueName string, message types.Message) error {
	_, err := s.instance().DeleteMessage(context.Background(), &sqs.DeleteMessageInput{
		QueueUrl:      s.MustGetURL(queueName),
		ReceiptHandle: message.ReceiptHandle,
	})
	return err
}

// BatchDeleteMessages 批量删除消息
func (s *SQS) BatchDeleteMessages(queueName string, messages []types.Message) (int, error) {
	entries := make([]types.DeleteMessageBatchRequestEntry, 0, len(messages))
	for _, message := range messages {
		id := message.MessageId
		if id == nil || *id == "" {
			id = aws.String(e2exec.Must(e2crypto.RandomString(16)))
		}
		entries = append(entries, types.DeleteMessageBatchRequestEntry{Id: id, ReceiptHandle: message.ReceiptHandle})
	}
	out, err := s.instance().DeleteMessageBatch(context.Background(), &sqs.DeleteMessageBatchInput{
		QueueUrl: s.MustGetURL(queueName),
		Entries:  entries,
	})
	if out == nil {
		return 0, err
	}
	return len(messages) - len(out.Failed), err
}

// GetQueueAttributes 根据队列名获取队列属性
func (s *SQS) GetQueueAttributes(queueName string) (map[string]string, error) {
	out, err := s.instance().GetQueueAttributes(context.Background(), &sqs.GetQueueAttributesInput{
		QueueUrl:       s.MustGetURL(queueName),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if out == nil {
		return nil, err
	}
	return out.Attributes, err
}

// GetQueueAttributesWithURL 根据队列URL获取队列属性
func (s *SQS) GetQueueAttributesWithURL(queueURL string) (map[string]string, error) {
	out, err := s.instance().GetQueueAttributes(context.Background(), &sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []types.QueueAttributeName{types.QueueAttributeNameAll},
	})
	if out == nil {
		return nil, err
	}
	return out.Attributes, err
}

// ListQueues 列出所有的队列 URL
func (s *SQS) ListQueues() ([]string, error) {
	out, err := s.instance().ListQueues(context.Background(), &sqs.ListQueuesInput{})
	if out == nil {
		return nil, err
	}
	return out.QueueUrls, err
}

// ListQueueNames 列出所有队列的名字
func (s *SQS) ListQueueNames() ([]string, error) {
	var names []string
	qs, err := s.ListQueues()
	if err != nil {
		return names, err
	}
	for _, q := range qs {
		names = append(names, filepath.Base(q))
	}
	return names, nil
}

// ParseSQSPath 解析出自定义的 sqs url,如 sqs://<queue_name>
func (s *SQS) ParseSQSPath(sqsPath string) (string, error) {
	if !strings.HasPrefix(sqsPath, "sqs://") {
		return "", fmt.Errorf(" illegal parameter %v", sqsPath)
	}
	u, err := url.Parse(sqsPath)
	if err != nil {
		return "", err
	}
	return u.Host, nil
}
