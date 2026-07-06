# S3 File Gateway

[Русская версия](README.ru.md)

Small HTTP gateway for working with S3-compatible object storage.

The service provides a thin REST API for uploading, downloading, and deleting files in S3 or MinIO. HTTP handlers take care of request validation, upload limits, streaming, response codes, and JSON errors. Storage-specific logic stays in a dedicated S3 adapter.

## ✨ What It Does

- Uploads files to an S3-compatible bucket via `multipart/form-data`.
- Downloads files by bucket and object key.
- Deletes objects by bucket and object key.
- Streams file content instead of loading whole files into memory.
- Preserves object `Content-Type` on download.
- Limits upload request size to protect the service from oversized bodies.
- Returns JSON error responses.
- Uses structured logging with `zap`.
- Exposes health, readiness, and Prometheus metrics endpoints.
- Runs locally with Docker Compose, MinIO, PostgreSQL, Prometheus, and Grafana.
- Supports local `.env` files for development and external environment variables for Docker.

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

### Readiness Check

```http
GET /readyz
```

Checks whether the gateway can reach the configured S3-compatible storage and access the readiness bucket.

> The current readiness bucket is intentionally fixed in code while the project is still evolving.

### Metrics

```http
GET /metrics
```

Exposes Prometheus metrics from the Go runtime, process collector, and HTTP handler.

Prometheus scrapes the gateway inside the Docker network at:

```text
app:8000
```

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
| `503 Service Unavailable` | Readiness check failed |

## ⚙️ Configuration

The application reads configuration from environment variables. A local `.env` file is loaded when present, but it is not required.

| Variable | Required | Description | Example |
| --- | --- | --- | --- |
| `HTTP_ADDR` | No | HTTP listen address | `0.0.0.0:8000` |
| `S3_ENDPOINT` | Yes | S3-compatible endpoint | `http://s3gw-minio:9000` |
| `S3_REGION` | Yes | S3 region | `us-east-1` |
| `S3_KEY` | Yes | Access key | `minioadmin` |
| `S3_SECRET` | Yes | Secret key | `minioadmin` |

Example `.env` for local Docker Compose:

```env
DB_USER=gateway
DB_PASSWORD=verysecret
DB_NAME=gateway

MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin

S3_KEY=minioadmin
S3_SECRET=minioadmin
S3_REGION=us-east-1
S3_ENDPOINT=http://s3gw-minio:9000

HTTP_ADDR=0.0.0.0:8000
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
| Gateway metrics | `http://localhost:8000/metrics` |
| MinIO S3 API | `http://localhost:9000` |
| MinIO Console | `http://localhost:9001` |
| Prometheus | `http://localhost:9090` |
| Grafana | `http://localhost:3000` |
| PostgreSQL | `localhost:5432` |

Create a bucket in MinIO before uploading files. You can do it through the MinIO Console or any S3-compatible client.

## 📊 Observability

The gateway exposes Prometheus metrics at `/metrics`.

The Compose stack includes:

- Prometheus scraping `app:8000`;
- Grafana for dashboards;
- standard Go runtime and process metrics out of the box.

Useful metrics available immediately:

| Metric | Meaning |
| --- | --- |
| `go_goroutines` | Current number of goroutines |
| `go_memstats_heap_alloc_bytes` | Heap memory currently allocated |
| `process_cpu_seconds_total` | CPU time consumed by the process |
| `process_resident_memory_bytes` | Resident memory used by the process |

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

Integration tests require a running S3-compatible backend.

To run them against local MinIO:

```bash
docker compose up -d minio
```

Then run:

```bash
S3_ENDPOINT=http://localhost:9000 \
S3_REGION=us-east-1 \
S3_KEY=minioadmin \
S3_SECRET=minioadmin \
go test ./s3_storage
```

On PowerShell:

```powershell
$env:S3_ENDPOINT="http://localhost:9000"
$env:S3_REGION="us-east-1"
$env:S3_KEY="minioadmin"
$env:S3_SECRET="minioadmin"
go test ./s3_storage
```

The integration test creates a temporary bucket, uploads an object, downloads it, deletes it, and verifies that the deleted key is no longer available.

## 🗂️ Project Structure

```text
cmd/
  main.go                   Application entry point and dependency wiring

config/
  config.go                 Environment-based configuration

router/
  router.go                 HTTP route registration
  handlers/
    healthz.go              Health endpoint
    readyz.go               Readiness endpoint
    get_file.go             Download handler
    put_file.go             Upload handler
    delete_file.go          Delete handler
    helpers.go              Shared HTTP response helpers
    handlers_test.go        Handler unit tests

s3_storage/
  storage.go                S3-compatible storage adapter
  interface.go              Storage interface used by handlers
  errors.go                 S3-specific error helpers
  storage_integration_test.go
                            MinIO/S3 integration test

Dockerfile                  Multi-stage image build
compose.yaml                Local stack: app, MinIO, PostgreSQL, Prometheus, Grafana
prometheus.yml              Prometheus scrape config
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
- The Docker image has a `/healthz` healthcheck.
- Prometheus and Grafana are included for local observability.

Potential next steps:

- authentication and authorization;
- request logging middleware;
- custom HTTP and S3 metrics;
- bucket creation workflow or bootstrap script;
- PostgreSQL-backed file metadata;
- more precise storage error mapping for upload failures.
