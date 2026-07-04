package s3_storage

import (
	"context"
	"io"
)

type S3Storage interface {
	GetObject(ctx context.Context, bucket, key string) (io.Reader, error)
	PutObject(ctx context.Context, bucket, key string, file io.ReadCloser, contentType string) error
	DeleteObject(ctx context.Context, bucket, key string) error
}
