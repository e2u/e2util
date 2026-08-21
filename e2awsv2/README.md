# e2awsv2

`e2awsv2` is a leftover stub. AWS SDK for Go v2 helpers now live in [`e2aws`](../e2aws).

Use:

```go
import "github.com/e2u/e2util/e2aws"
import "github.com/e2u/e2util/e2aws/e2s3"
```

`e2awsv2/e2s3v2` still exposes `ParseS3Path` for compatibility. New code should call `e2s3.S3.ParseS3Path` instead.
