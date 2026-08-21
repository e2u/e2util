package e2s3v2

import "testing"

func TestParseS3Path(t *testing.T) {
	s := &S3{}
	bucket, key, err := s.ParseS3Path("s3://my-bucket/path/to/file.txt")
	if err != nil {
		t.Fatal(err)
	}
	if bucket != "my-bucket" || key != "path/to/file.txt" {
		t.Errorf("ParseS3Path = %q, %q", bucket, key)
	}

	bucket, key, err = s.ParseS3Path("s3://only-bucket")
	if err != nil {
		t.Fatal(err)
	}
	if bucket != "only-bucket" {
		t.Errorf("bucket = %q", bucket)
	}

	if _, _, err := s.ParseS3Path("https://example.com/x"); err == nil {
		t.Fatal("expected error for non-s3 path")
	}
}
