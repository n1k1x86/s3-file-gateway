package s3_storage

import (
	"context"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsv4 "github.com/aws/aws-sdk-go-v2/aws/signer/v4"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func TestS3StorageIntegrationPutGetDelete(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	endpoint := envOrDefault("S3_ENDPOINT", "http://localhost:9000")
	region := envOrDefault("S3_REGION", "us-east-1")
	key := envOrDefault("S3_KEY", "minioadmin")
	secret := envOrDefault("S3_SECRET", "minioadmin")
	bucket := envOrDefault("S3_TEST_BUCKET", "s3-file-gateway-it-"+time.Now().UTC().Format("20060102150405"))
	objectKey := "integration/file.txt"
	content := "hello from integration test"
	contentType := "text/plain"

	adminClient, err := newIntegrationS3Client(ctx, key, secret, region, endpoint)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := adminClient.CreateBucket(ctx, &s3.CreateBucketInput{
		Bucket: aws.String(bucket),
	}); err != nil {
		t.Fatalf("create bucket: %v", err)
	}

	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cleanupCancel()

		_, _ = adminClient.DeleteObject(cleanupCtx, &s3.DeleteObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(objectKey),
		})
		_, _ = adminClient.DeleteBucket(cleanupCtx, &s3.DeleteBucketInput{
			Bucket: aws.String(bucket),
		})
	})

	storage, err := NewS3Storage(ctx, key, secret, region, endpoint)
	if err != nil {
		t.Fatal(err)
	}

	if err := storage.PutObject(ctx, bucket, objectKey, io.NopCloser(strings.NewReader(content)), contentType, int64(len(content))); err != nil {
		t.Fatalf("put object: %v", err)
	}

	reader, gotContentType, err := storage.GetObject(ctx, bucket, objectKey)
	if err != nil {
		t.Fatalf("get object: %v", err)
	}
	defer reader.Close()

	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read object body: %v", err)
	}
	if string(data) != content {
		t.Fatalf("object body = %q, want %q", string(data), content)
	}
	if gotContentType != contentType {
		t.Fatalf("content type = %q, want %q", gotContentType, contentType)
	}

	if err := storage.DeleteObject(ctx, bucket, objectKey); err != nil {
		t.Fatalf("delete object: %v", err)
	}

	reader, _, err = storage.GetObject(ctx, bucket, objectKey)
	if reader != nil {
		_ = reader.Close()
	}
	if err == nil {
		t.Fatal("expected get deleted object to fail")
	}
	if !IsErrNoSuchKey(err) {
		t.Fatalf("expected no such key error, got %T: %v", err, err)
	}
}

func newIntegrationS3Client(ctx context.Context, key, secret, region, endpoint string) (*s3.Client, error) {
	cfg, err := awsconfig.LoadDefaultConfig(ctx,
		awsconfig.WithRegion(region),
		awsconfig.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(key, secret, ""),
		),
	)
	if err != nil {
		return nil, err
	}

	return s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
		o.UsePathStyle = true
		o.RequestChecksumCalculation = aws.RequestChecksumCalculationWhenRequired
		o.APIOptions = append(o.APIOptions, awsv4.SwapComputePayloadSHA256ForUnsignedPayloadMiddleware)
	}), nil
}

func envOrDefault(name, defaultValue string) string {
	value := os.Getenv(name)
	if value == "" {
		return defaultValue
	}
	return value
}
