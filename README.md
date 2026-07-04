# S3 File Gateway

HTTP gateway for working with S3-compatible object storage.

The service exposes a small REST API for uploading, downloading, and deleting files in S3 or MinIO. It is intentionally thin: HTTP handlers validate request parameters, apply upload limits, stream file data, and delegate storage operations to a dedicated S3 adapter.

## ✨ What It Does

- Uploads files to an S3-compatible bucket via `multipart/form-data`.
- Downloads files by bucket and object key.
- Deletes objects by bucket and object key.
- Streams file content instead of loading whole files into memory.
- Preserves object `Content-Type` on download.
- Limits upload request size to protect the service from oversized bodies.
- Returns JSON error responses.
- Uses structured logging with `zap`.
- Supports local `.env` files for development and runtime environment variables for Docker.
- Runs with Docker Compose using MinIO as the S3-compatible backend.

## 🔌 API

The object key is passed as a query parameter so keys can contain slashes:

```text
photos/2026/july/image.jpg
```

### Health Check

```http
GET /healthz
```

Returns `200 OK` when the HTTP server is alive.

### Upload File

```http
POST /files/{bucket}?key={object-key}
Content-Type: multipart/form-data
```

Form field:

```text
file
```

Example:

```bash
curl -i \
  -F "file=@./photo.jpg" \
  "http://localhost:8000/files/my-bucket?key=photos/2026/photo.jpg"
```

Successful response:

```http
HTTP/1.1 201 Created
Content-Type: application/json
```

```json
{
  "key": "photos/2026/photo.jpg"
}
```

Upload requests are limited to `50 MB` plus multipart overhead.

### Download File

```http
GET /files/{bucket}?key={object-key}
```

Example:

```bash
curl -o photo.jpg \
  "http://localhost:8000/files/my-bucket?key=photos/2026/photo.jpg"
```

Successful response:

```http
HTTP/1.1 200 OK
Content-Type: image/jpeg
```

The response body is streamed from S3 to the client.

### Delete File

```http
DELETE /files/{bucket}?key={object-key}
```

Example:

```bash
curl -i -X DELETE \
  "http://localhost:8000/files/my-bucket?key=photos/2026/photo.jpg"
```

Successful response:

```http
HTTP/1.1 204 No Content
```

## ⚠️ Error Format

Errors are returned as JSON:

```json
{
  "err": "key is empty"
}
```

Common statuses:

| Status | Meaning |
| --- | --- |
| `400 Bad Request` | Missing `key`, invalid multipart body, or invalid request data |
| `404 Not Found` | Bucket or object was not found in S3 |
| `413 Payload Too Large` | Upload request exceeded the configured size limit |
| `500 Internal Server Error` | Unexpected storage or streaming error |

## ⚙️ Configuration

The application reads configuration from environment variables. A local `.env` file is loaded when present, but it is not required.

| Variable | Required | Description | Example |
| --- | --- | --- | --- |
| `HTTP_ADDR` | No | HTTP listen address | `0.0.0.0:8000` |
| `S3_ENDPOINT` | Yes | S3-compatible endpoint | `http://minio:9000` |
| `S3_REGION` | Yes | S3 region | `us-east-1` |
| `S3_KEY` | Yes | Access key | `minioadmin` |
| `S3_SECRET` | Yes | Secret key | `minioadmin` |

Example `.env` for local Docker Compose:

```env
HTTP_ADDR=0.0.0.0:8000
S3_ENDPOINT=http://minio:9000
S3_REGION=us-east-1
S3_KEY=minioadmin
S3_SECRET=minioadmin

MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin
```

The `.env` file is passed to the container by Compose and is excluded from the Docker build context by `.dockerignore`.

## 🚀 Running Locally

Start the stack:

```bash
docker compose up --build
```

Services:

| Service | URL |
| --- | --- |
| Gateway | `http://localhost:8000` |
| MinIO S3 API | `http://localhost:9000` |
| MinIO Console | `http://localhost:9001` |

Create a bucket in MinIO before uploading files. You can do it through the MinIO Console or any S3-compatible client.

## 🧰 Running Without Docker

Set the required environment variables and run:

```bash
go run ./cmd/main.go
```

The service defaults to:

```text
0.0.0.0:8000
```

when `HTTP_ADDR` is not set.

## ✅ Tests

Run all tests:

```bash
go test ./...
```

The current test suite covers HTTP handler behavior with a fake storage implementation:

- successful upload;
- successful download;
- successful delete;
- S3 not-found mapping;
- response body closing after download.

Integration tests are disabled by default because they require a running S3-compatible backend.

To run them against local MinIO:

```bash
docker compose up -d minio
```

Then create a bucket or let the test create a temporary one, and run:

```bash
S3_INTEGRATION_TESTS=1 \
S3_ENDPOINT=http://localhost:9000 \
S3_REGION=us-east-1 \
S3_KEY=minioadmin \
S3_SECRET=minioadmin \
go test ./s3_storage
```

On PowerShell:

```powershell
$env:S3_INTEGRATION_TESTS="1"
$env:S3_ENDPOINT="http://localhost:9000"
$env:S3_REGION="us-east-1"
$env:S3_KEY="minioadmin"
$env:S3_SECRET="minioadmin"
go test ./s3_storage
```

## 🗂️ Project Structure

```text
cmd/
  main.go                 Application entry point and dependency wiring

config/
  config.go               Environment-based configuration

router/
  router.go               HTTP route registration
  healthz.go              Health endpoint
  handlers/               HTTP handlers and handler tests

s3_storage/
  storage.go              S3-compatible storage adapter
  interface.go            Storage interface used by handlers
  errors.go               S3-specific error helpers
```

## 🧠 Design Notes

The gateway keeps the storage adapter separate from the HTTP layer:

```text
HTTP client
  -> router handlers
  -> S3Storage interface
  -> AWS SDK / MinIO / S3-compatible storage
```

Handlers are responsible for HTTP concerns: path/query parameters, multipart parsing, response codes, and JSON errors.

The `s3_storage` package is responsible for S3 concerns: AWS SDK client setup, path-style addressing, object streaming, content type handling, and S3-specific error checks.

This keeps the codebase small, testable, and easy to extend without adding unnecessary layers.

## 🔐 Security And Operational Notes

- `.env` is not copied into the Docker image.
- Upload body size is limited at the HTTP layer.
- Files are streamed to and from storage instead of being fully buffered in memory.
- The final Docker image runs as a non-root user.
- Logs are structured JSON logs through `zap`.

Potential next steps:

- authentication and authorization;
- bucket creation workflow or bootstrap script;
- integration tests with MinIO;
- request logging middleware;
- metrics and tracing;
- more precise storage error mapping for upload failures.
