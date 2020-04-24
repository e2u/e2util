package e2sqs

import (
	"fmt"
	"net/url"
	"path/filepath"
	"strings"
	"sync"

	"git.panda-fintech.com/golang/e2util/e2crypto"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

type SQS struct {
	urlCache sync.Map
	sess     *session.Session
}

func New(sess *session.Session, cfgs ...*aws.Config) *SQS {
	sess = sess.Copy(cfgs...)
	return &SQS{
		sess: sess,
	}
}

func (s *SQS) instance() *sqs.SQS {
	return sqs.New(s.sess)
}

func (s *SQS) GetURL(queueName string) (*string, error) {
	if v, ok := s.urlCache.Load(queueName); ok {
		return v.(*string), nil
	}
	out, err := s.instance().GetQueueUrl(&sqs.GetQueueUrlInput{QueueName: aws.String(queueName)})
	if err == nil {
		s.urlCache.Store(queueName, out.QueueUrl)
	}
	return out.QueueUrl, err
}

func (s *SQS) MustGetURL(queueName string) *string {
	url, _ := s.GetURL(queueName)
	return url
}

// SendMessage 发送单条消息到队列
func (s *SQS) SendMessage(queueName string, message string) error {
	_, err := s.instance().SendMessage(&sqs.SendMessageInput{
		QueueUrl:    s.MustGetURL(queueName),
		MessageBody: aws.String(message),
	})
	return err
}

// BatchSendMessages 批量发送消息到队列
func (s *SQS) BatchSendMessages(queueName string, messages []string) (int, error) {
	out, err := s.instance().SendMessageBatch(&sqs.SendMessageBatchInput{
		QueueUrl: s.MustGetURL(queueName),
		Entries: func() []*sqs.SendMessageBatchRequestEntry {
			var re []*sqs.SendMessageBatchRequestEntry
			for _, message := range messages {
				re = append(re, &sqs.SendMessageBatchRequestEntry{MessageBody: aws.String(message), Id: aws.String(e2crypto.RandomString(16))})
			}
			return re
		}(),
	})
	return len(messages) - len(out.Failed), err
}

// ReceiveMessage 接收队列的消息
func (s *SQS) ReceiveMessage(queueName string, maxNumber int64) ([]*sqs.Message, error) {
	out, err := s.instance().ReceiveMessage(&sqs.ReceiveMessageInput{
		QueueUrl:            s.MustGetURL(queueName),
		MaxNumberOfMessages: aws.Int64(maxNumber),
	})
	return out.Messages, err
}

// DeleteMessage 删除单条消息
func (s *SQS) DeleteMessage(queueName string, message *sqs.Message) error {
	_, err := s.instance().DeleteMessage(&sqs.DeleteMessageInput{
		QueueUrl:      s.MustGetURL(queueName),
		ReceiptHandle: message.ReceiptHandle,
	})
	return err
}

// BatchDeleteMessages 批量删除消息
func (s *SQS) BatchDeleteMessages(queueName string, messages []*sqs.Message) (int, error) {
	out, err := s.instance().DeleteMessageBatch(&sqs.DeleteMessageBatchInput{
		QueueUrl: s.MustGetURL(queueName),
		Entries: func() []*sqs.DeleteMessageBatchRequestEntry {
			var re []*sqs.DeleteMessageBatchRequestEntry
			for _, message := range messages {
				re = append(re, &sqs.DeleteMessageBatchRequestEntry{ReceiptHandle: message.ReceiptHandle})
			}
			return re
		}(),
	})
	return len(messages) - len(out.Failed), err
}

// GetQueueAttributes 根据队列名获取队列属性
func (s *SQS) GetQueueAttributes(queueName string) (map[string]*string, error) {
	out, err := s.instance().GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl:       s.MustGetURL(queueName),
		AttributeNames: []*string{aws.String(sqs.QueueAttributeNameAll)},
	})

	return out.Attributes, err
}

// GetQueueAttributesWithURL 根据队列URL获取队列属性
func (s *SQS) GetQueueAttributesWithURL(queueURL string) (map[string]*string, error) {
	out, err := s.instance().GetQueueAttributes(&sqs.GetQueueAttributesInput{
		QueueUrl:       aws.String(queueURL),
		AttributeNames: []*string{aws.String(sqs.QueueAttributeNameAll)},
	})

	return out.Attributes, err
}

// ListQueues 列出所有的队列 URL
func (s *SQS) ListQueues() ([]*string, error) {
	out, err := s.instance().ListQueues(&sqs.ListQueuesInput{})
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
		names = append(names, filepath.Base(aws.StringValue(q)))
	}
	return names, nil
}

// ParseSQSUrl 解析出自定义的 sqs url,如 sqs://<queue_name>
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
