# e2aws

## Overview
The `e2aws` module provides AWS SDK for Go v2 helpers for S3, DynamoDB, and SQS.

The old `e2awsv2` package is a leftover stub. New code should import `e2aws` (and `e2aws/e2s3`, `e2aws/e2dynamodb`, `e2aws/e2sqs`).

## Installation
```bash
go get github.com/e2u/e2util/e2aws
```

## Usage
```go
import "github.com/e2u/e2util/e2aws"

// Example: create an S3 client
client := e2aws.NewS3Client(cfg)
```

## Examples
*Add example usage for S3 upload/download.*

## API Reference
*List exported symbols.*
