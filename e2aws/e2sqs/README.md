# e2sqs

AWS SDK Go v2 的 SQS 封裝：送訊、收訊、刪除、列隊、屬性。

SQS helpers on AWS SDK for Go v2: send, receive, delete, list, attributes.

## 安裝 / Installation

```bash
go get github.com/e2u/e2util/e2aws/e2sqs
```

## 功能 / Features

- **URL**：`GetURL`、`MustGetURL`、`ParseSQSPath`
- **訊息 / Messages**：`SendMessage`、`BatchSendMessages`、`ReceiveMessage`
- **刪除 / Delete**：`DeleteMessage`、`BatchDeleteMessages`
- **隊列 / Queues**：`ListQueues`、`ListQueueNames`、`GetQueueAttributes`

## 用法 / Usage

```go
import (
    "github.com/e2u/e2util/e2aws"
    "github.com/e2u/e2util/e2aws/e2sqs"
)

cfg := e2aws.NewSession("us-east-1")
q := e2sqs.New(cfg)
_, err := q.SendMessage("my-queue", "hello")
```
