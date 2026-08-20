package e2sqs

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testQueueURL = "https://sqs.us-east-1.amazonaws.com/123456789012/orders"

func testConfig() aws.Config {
	return aws.Config{
		Region:      "us-east-1",
		Credentials: credentials.NewStaticCredentialsProvider("AKID", "SECRET", ""),
	}
}

func writeJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/x-amz-json-1.0")
	w.WriteHeader(status)
	_, _ = io.WriteString(w, body)
}

func newTestSQS(t *testing.T, handler http.HandlerFunc) *SQS {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return New(testConfig(), func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(srv.URL)
		o.HTTPClient = srv.Client()
		o.Retryer = aws.NopRetryer{}
	})
}

func queueURLHandler(t *testing.T, extra func(w http.ResponseWriter, r *http.Request, target string, body []byte)) http.HandlerFunc {
	t.Helper()
	return func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-Amz-Target")
		body, _ := io.ReadAll(r.Body)
		if target == "AmazonSQS.GetQueueUrl" {
			writeJSON(w, http.StatusOK, `{"QueueUrl":"`+testQueueURL+`"}`)
			return
		}
		if extra != nil {
			extra(w, r, target, body)
			return
		}
		writeJSON(w, http.StatusOK, `{}`)
	}
}

func TestNew(t *testing.T) {
	s := New(testConfig(), func(o *sqs.Options) {
		o.Region = "us-west-2"
	})
	require.NotNil(t, s)
	assert.NotNil(t, s.client)
	assert.Equal(t, s.client, s.instance())
	assert.Equal(t, "us-west-2", s.client.Options().Region)
}

func TestParseSQSPath(t *testing.T) {
	s := &SQS{}

	name, err := s.ParseSQSPath("sqs://orders")
	require.NoError(t, err)
	assert.Equal(t, "orders", name)

	_, err = s.ParseSQSPath("https://sqs.us-east-1.amazonaws.com/123/orders")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "illegal parameter")

	_, err = s.ParseSQSPath("sqs://[::1")
	require.Error(t, err)
}

func TestGetURL(t *testing.T) {
	t.Run("fetches and caches url", func(t *testing.T) {
		var calls atomic.Int32
		s := newTestSQS(t, func(w http.ResponseWriter, r *http.Request) {
			calls.Add(1)
			assert.Equal(t, "AmazonSQS.GetQueueUrl", r.Header.Get("X-Amz-Target"))
			var payload map[string]any
			require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
			assert.Equal(t, "orders", payload["QueueName"])
			writeJSON(w, http.StatusOK, `{"QueueUrl":"`+testQueueURL+`"}`)
		})

		u, err := s.GetURL("orders")
		require.NoError(t, err)
		assert.Equal(t, testQueueURL, aws.ToString(u))

		u2, err := s.GetURL("orders")
		require.NoError(t, err)
		assert.Equal(t, u, u2)
		assert.Equal(t, int32(1), calls.Load())
	})

	t.Run("returns cached value without request", func(t *testing.T) {
		s := New(testConfig())
		want := aws.String(testQueueURL)
		s.urlCache.Store("orders", want)
		got, err := s.GetURL("orders")
		require.NoError(t, err)
		assert.Equal(t, want, got)
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestSQS(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"QueueDoesNotExist","message":"missing"}`)
		})
		u, err := s.GetURL("missing")
		require.Error(t, err)
		assert.Nil(t, u)
	})
}

func TestMustGetURL(t *testing.T) {
	t.Run("returns url", func(t *testing.T) {
		s := newTestSQS(t, queueURLHandler(t, nil))
		assert.Equal(t, testQueueURL, aws.ToString(s.MustGetURL("orders")))
	})

	t.Run("returns nil on error", func(t *testing.T) {
		s := newTestSQS(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"QueueDoesNotExist","message":"missing"}`)
		})
		assert.Nil(t, s.MustGetURL("missing"))
	})
}

func TestSendMessage(t *testing.T) {
	t.Run("sends body to resolved queue", func(t *testing.T) {
		var got map[string]any
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			assert.Equal(t, "AmazonSQS.SendMessage", target)
			require.NoError(t, json.Unmarshal(body, &got))
			writeJSON(w, http.StatusOK, `{"MessageId":"mid"}`)
		}))
		err := s.SendMessage("orders", "hello")
		require.NoError(t, err)
		assert.Equal(t, testQueueURL, got["QueueUrl"])
		assert.Equal(t, "hello", got["MessageBody"])
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"InvalidMessageContents","message":"bad"}`)
		}))
		err := s.SendMessage("orders", "hello")
		require.Error(t, err)
	})
}

func TestBatchSendMessages(t *testing.T) {
	t.Run("counts successful messages", func(t *testing.T) {
		var got map[string]any
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			assert.Equal(t, "AmazonSQS.SendMessageBatch", target)
			require.NoError(t, json.Unmarshal(body, &got))
			writeJSON(w, http.StatusOK, `{"Successful":[{"Id":"1","MessageId":"m1","MD5OfMessageBody":"d"}],"Failed":[{"Id":"2","Code":"InternalError","SenderFault":false,"Message":"fail"}]}`)
		}))
		n, err := s.BatchSendMessages("orders", []string{"a", "b"})
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Equal(t, testQueueURL, got["QueueUrl"])
		entries := got["Entries"].([]any)
		assert.Len(t, entries, 2)
		first := entries[0].(map[string]any)
		assert.Equal(t, "a", first["MessageBody"])
		assert.NotEmpty(t, first["Id"])
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"EmptyBatchRequest","message":"empty"}`)
		}))
		n, err := s.BatchSendMessages("orders", []string{"a"})
		require.Error(t, err)
		assert.Equal(t, 0, n)
	})
}

func TestReceiveMessage(t *testing.T) {
	t.Run("returns messages", func(t *testing.T) {
		var got map[string]any
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			assert.Equal(t, "AmazonSQS.ReceiveMessage", target)
			require.NoError(t, json.Unmarshal(body, &got))
			writeJSON(w, http.StatusOK, `{"Messages":[{"MessageId":"m1","ReceiptHandle":"rh1","Body":"hello"}]}`)
		}))
		msgs, err := s.ReceiveMessage("orders", 5)
		require.NoError(t, err)
		require.Len(t, msgs, 1)
		assert.Equal(t, "hello", aws.ToString(msgs[0].Body))
		assert.Equal(t, "rh1", aws.ToString(msgs[0].ReceiptHandle))
		assert.Equal(t, testQueueURL, got["QueueUrl"])
		assert.EqualValues(t, 5, got["MaxNumberOfMessages"])
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"OverLimit","message":"too many"}`)
		}))
		msgs, err := s.ReceiveMessage("orders", 1)
		require.Error(t, err)
		assert.Nil(t, msgs)
	})
}

func TestDeleteMessage(t *testing.T) {
	t.Run("deletes by receipt handle", func(t *testing.T) {
		var got map[string]any
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			assert.Equal(t, "AmazonSQS.DeleteMessage", target)
			require.NoError(t, json.Unmarshal(body, &got))
			writeJSON(w, http.StatusOK, `{}`)
		}))
		err := s.DeleteMessage("orders", types.Message{ReceiptHandle: aws.String("rh1")})
		require.NoError(t, err)
		assert.Equal(t, testQueueURL, got["QueueUrl"])
		assert.Equal(t, "rh1", got["ReceiptHandle"])
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"ReceiptHandleIsInvalid","message":"bad"}`)
		}))
		err := s.DeleteMessage("orders", types.Message{ReceiptHandle: aws.String("bad")})
		require.Error(t, err)
	})
}

func TestBatchDeleteMessages(t *testing.T) {
	t.Run("uses message id and generates missing ids", func(t *testing.T) {
		var got map[string]any
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			assert.Equal(t, "AmazonSQS.DeleteMessageBatch", target)
			require.NoError(t, json.Unmarshal(body, &got))
			writeJSON(w, http.StatusOK, `{"Successful":[{"Id":"m1"}],"Failed":[{"Id":"x","Code":"ReceiptHandleIsInvalid","SenderFault":true,"Message":"bad"}]}`)
		}))
		n, err := s.BatchDeleteMessages("orders", []types.Message{
			{MessageId: aws.String("m1"), ReceiptHandle: aws.String("rh1")},
			{MessageId: aws.String(""), ReceiptHandle: aws.String("rh2")},
		})
		require.NoError(t, err)
		assert.Equal(t, 1, n)
		entries := got["Entries"].([]any)
		require.Len(t, entries, 2)
		assert.Equal(t, "m1", entries[0].(map[string]any)["Id"])
		assert.Equal(t, "rh1", entries[0].(map[string]any)["ReceiptHandle"])
		assert.NotEmpty(t, entries[1].(map[string]any)["Id"])
		assert.Equal(t, "rh2", entries[1].(map[string]any)["ReceiptHandle"])
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"EmptyBatchRequest","message":"empty"}`)
		}))
		n, err := s.BatchDeleteMessages("orders", []types.Message{{ReceiptHandle: aws.String("rh")}})
		require.Error(t, err)
		assert.Equal(t, 0, n)
	})
}

func TestGetQueueAttributes(t *testing.T) {
	t.Run("by queue name", func(t *testing.T) {
		var got map[string]any
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			assert.Equal(t, "AmazonSQS.GetQueueAttributes", target)
			require.NoError(t, json.Unmarshal(body, &got))
			writeJSON(w, http.StatusOK, `{"Attributes":{"ApproximateNumberOfMessages":"3","QueueArn":"arn:aws:sqs:us-east-1:123:orders"}}`)
		}))
		attrs, err := s.GetQueueAttributes("orders")
		require.NoError(t, err)
		assert.Equal(t, "3", attrs["ApproximateNumberOfMessages"])
		assert.Equal(t, testQueueURL, got["QueueUrl"])
		assert.Equal(t, []any{"All"}, got["AttributeNames"])
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestSQS(t, queueURLHandler(t, func(w http.ResponseWriter, r *http.Request, target string, body []byte) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"QueueDoesNotExist","message":"missing"}`)
		}))
		attrs, err := s.GetQueueAttributes("orders")
		require.Error(t, err)
		assert.Nil(t, attrs)
	})
}

func TestGetQueueAttributesWithURL(t *testing.T) {
	t.Run("by url", func(t *testing.T) {
		var got map[string]any
		s := newTestSQS(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "AmazonSQS.GetQueueAttributes", r.Header.Get("X-Amz-Target"))
			require.NoError(t, json.NewDecoder(r.Body).Decode(&got))
			writeJSON(w, http.StatusOK, `{"Attributes":{"VisibilityTimeout":"30"}}`)
		})
		attrs, err := s.GetQueueAttributesWithURL(testQueueURL)
		require.NoError(t, err)
		assert.Equal(t, "30", attrs["VisibilityTimeout"])
		assert.Equal(t, testQueueURL, got["QueueUrl"])
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestSQS(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"InvalidAddress","message":"bad url"}`)
		})
		attrs, err := s.GetQueueAttributesWithURL("bad")
		require.Error(t, err)
		assert.Nil(t, attrs)
	})
}

func TestListQueues(t *testing.T) {
	t.Run("returns urls", func(t *testing.T) {
		s := newTestSQS(t, func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, "AmazonSQS.ListQueues", r.Header.Get("X-Amz-Target"))
			writeJSON(w, http.StatusOK, `{"QueueUrls":["`+testQueueURL+`","https://sqs.us-east-1.amazonaws.com/123456789012/payments"]}`)
		})
		urls, err := s.ListQueues()
		require.NoError(t, err)
		assert.Equal(t, []string{testQueueURL, "https://sqs.us-east-1.amazonaws.com/123456789012/payments"}, urls)
	})

	t.Run("api error", func(t *testing.T) {
		s := newTestSQS(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"InvalidAddress","message":"bad"}`)
		})
		urls, err := s.ListQueues()
		require.Error(t, err)
		assert.Nil(t, urls)
	})
}

func TestListQueueNames(t *testing.T) {
	t.Run("extracts names from urls", func(t *testing.T) {
		s := newTestSQS(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusOK, `{"QueueUrls":["`+testQueueURL+`","https://sqs.us-east-1.amazonaws.com/123456789012/payments"]}`)
		})
		names, err := s.ListQueueNames()
		require.NoError(t, err)
		assert.Equal(t, []string{"orders", "payments"}, names)
	})

	t.Run("propagates list error", func(t *testing.T) {
		s := newTestSQS(t, func(w http.ResponseWriter, r *http.Request) {
			writeJSON(w, http.StatusBadRequest, `{"__type":"InvalidAddress","message":"bad"}`)
		})
		names, err := s.ListQueueNames()
		require.Error(t, err)
		assert.Nil(t, names)
	})
}
