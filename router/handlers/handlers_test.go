package handlers

import (
	"bytes"
	"context"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

type fakeStorage struct {
	getObject func(ctx context.Context, bucket, key string) (io.ReadCloser, string, error)
	putObject func(ctx context.Context, bucket, key string, file io.ReadCloser, contentType string) error
	delObject func(ctx context.Context, bucket, key string) error
}

func (s fakeStorage) GetObject(ctx context.Context, bucket, key string) (io.ReadCloser, string, error) {
	return s.getObject(ctx, bucket, key)
}

func (s fakeStorage) PutObject(ctx context.Context, bucket, key string, file io.ReadCloser, contentType string) error {
	return s.putObject(ctx, bucket, key, file, contentType)
}

func (s fakeStorage) DeleteObject(ctx context.Context, bucket, key string) error {
	return s.delObject(ctx, bucket, key)
}

type closeTracker struct {
	*bytes.Reader
	closed bool
}

func (r *closeTracker) Close() error {
	r.closed = true
	return nil
}

func TestGetFileSuccess(t *testing.T) {
	body := &closeTracker{Reader: bytes.NewReader([]byte("hello"))}
	storage := fakeStorage{
		getObject: func(ctx context.Context, bucket, key string) (io.ReadCloser, string, error) {
			if bucket != "bucket" {
				t.Fatalf("bucket = %q, want %q", bucket, "bucket")
			}
			if key != "dir/file.txt" {
				t.Fatalf("key = %q, want %q", key, "dir/file.txt")
			}
			return body, "text/plain", nil
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/files/bucket?key=dir/file.txt", nil)
	req.SetPathValue("bucket", "bucket")
	rr := httptest.NewRecorder()

	GetFile(storage).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusOK)
	}
	if got := rr.Body.String(); got != "hello" {
		t.Fatalf("body = %q, want %q", got, "hello")
	}
	if got := rr.Header().Get("Content-Type"); got != "text/plain" {
		t.Fatalf("Content-Type = %q, want %q", got, "text/plain")
	}
	if !body.closed {
		t.Fatal("expected response body to be closed")
	}
}

func TestGetFileNotFound(t *testing.T) {
	storage := fakeStorage{
		getObject: func(ctx context.Context, bucket, key string) (io.ReadCloser, string, error) {
			return nil, "", &types.NoSuchKey{}
		},
	}

	req := httptest.NewRequest(http.MethodGet, "/files/bucket?key=missing.txt", nil)
	req.SetPathValue("bucket", "bucket")
	rr := httptest.NewRecorder()

	GetFile(storage).ServeHTTP(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNotFound)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("Content-Type = %q, want %q", got, "application/json")
	}
}

func TestPutFileSuccess(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreateFormFile("file", "upload.txt")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write([]byte("uploaded")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}

	storage := fakeStorage{
		putObject: func(ctx context.Context, bucket, key string, file io.ReadCloser, contentType string) error {
			if bucket != "bucket" {
				t.Fatalf("bucket = %q, want %q", bucket, "bucket")
			}
			if key != "dir/upload.txt" {
				t.Fatalf("key = %q, want %q", key, "dir/upload.txt")
			}
			data, err := io.ReadAll(file)
			if err != nil {
				t.Fatal(err)
			}
			if string(data) != "uploaded" {
				t.Fatalf("uploaded body = %q, want %q", string(data), "uploaded")
			}
			if contentType == "" {
				t.Fatal("expected content type to be passed")
			}
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodPost, "/files/bucket?key=dir/upload.txt", &body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.SetPathValue("bucket", "bucket")
	rr := httptest.NewRecorder()

	PutFile(storage).ServeHTTP(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusCreated)
	}
}

func TestDeleteFileSuccess(t *testing.T) {
	storage := fakeStorage{
		delObject: func(ctx context.Context, bucket, key string) error {
			if bucket != "bucket" {
				t.Fatalf("bucket = %q, want %q", bucket, "bucket")
			}
			if key != "dir/file.txt" {
				t.Fatalf("key = %q, want %q", key, "dir/file.txt")
			}
			return nil
		},
	}

	req := httptest.NewRequest(http.MethodDelete, "/files/bucket?key=dir/file.txt", nil)
	req.SetPathValue("bucket", "bucket")
	rr := httptest.NewRecorder()

	DeleteFile(storage).ServeHTTP(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rr.Code, http.StatusNoContent)
	}
}
