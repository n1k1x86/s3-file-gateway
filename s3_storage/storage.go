package s3_storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const (
	contextTimeout = time.Second * 10
)

type s3Storage struct {
	client *s3.Client
}

func (s *s3Storage) GetObject(ctx context.Context, bucket, key string) (io.Reader, error) {
	o, err := s.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return nil, err
	}

	if o == nil {
		return nil, fmt.Errorf("aws object is nil\n")
	}

	if o.Body == nil {
		return nil, fmt.Errorf("aws object body is nil\n")
	}

	return o.Body, nil
}

func (s *s3Storage) PutObject(ctx context.Context, bucket, key string, file io.ReadCloser, contentType string) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *s3Storage) DeleteObject(ctx context.Context, bucket, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return err
	}
	return nil
}

func NewS3Storage(ctx context.Context, key, secret, region, endpoint string) (S3Storage, error) {
	s3Ctx, s3Cancel := context.WithTimeout(ctx, contextTimeout)
	defer s3Cancel()

	cfg, err := config.LoadDefaultConfig(s3Ctx,
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				key,
				secret,
				"",
			),
		),
	)

	if err != nil {
		return nil, err
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = &endpoint
		o.UsePathStyle = true
	})

	return &s3Storage{
		client: client,
	}, nil
}
