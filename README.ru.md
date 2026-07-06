# S3 File Gateway

[English version](README.md)

Небольшой HTTP-шлюз для работы с S3-совместимым объектным хранилищем.

Сервис предоставляет тонкий REST API для загрузки, скачивания и удаления файлов в S3 или MinIO. HTTP-обработчики отвечают за валидацию запросов, ограничения размера upload, streaming, HTTP-коды и JSON-ошибки. Логика работы с хранилищем вынесена в отдельный S3 adapter.

## ✨ Что Умеет

- Загружает файлы в S3-совместимый bucket через `multipart/form-data`.
- Скачивает файлы по bucket и object key.
- Удаляет объекты по bucket и object key.
- Стримит содержимое файлов без полной загрузки в память.
- Сохраняет `Content-Type` объекта при скачивании.
- Ограничивает размер upload-запроса для защиты сервиса от слишком больших тел.
- Возвращает ошибки в JSON-формате.
- Использует структурированные логи через `zap`.
- Отдаёт health, readiness и Prometheus metrics endpoints.
- Локально запускается через Docker Compose вместе с MinIO, PostgreSQL, Prometheus и Grafana.
- Поддерживает локальный `.env` для разработки и внешние переменные окружения для Docker.

## 🔌 API

Object key передаётся query-параметром, чтобы в ключах можно было использовать слэши:

```text
photos/2026/july/image.jpg
```

### Health Check

```http
GET /healthz
```

Возвращает `200 OK`, если HTTP-сервер жив.

### Readiness Check

```http
GET /readyz
```

Проверяет, что gateway может обратиться к S3-совместимому хранилищу и получить доступ к readiness bucket.

> Сейчас readiness bucket намеренно зафиксирован в коде, пока проект развивается.

### Metrics

```http
GET /metrics
```

Отдаёт Prometheus-метрики Go runtime, process collector и HTTP handler.

Prometheus внутри Docker network ходит до gateway по адресу:

```text
app:8000
```

### Upload File

```http
POST /files/{bucket}?key={object-key}
Content-Type: multipart/form-data
```

Поле формы:

```text
file
```

Пример:

```bash
curl -i \
  -F "file=@./photo.jpg" \
  "http://localhost:8000/files/my-bucket?key=photos/2026/photo.jpg"
```

Успешный ответ:

```http
HTTP/1.1 201 Created
Content-Type: application/json
```

```json
{
  "key": "photos/2026/photo.jpg"
}
```

Upload-запросы ограничены `50 MB` плюс multipart overhead.

### Download File

```http
GET /files/{bucket}?key={object-key}
```

Пример:

```bash
curl -o photo.jpg \
  "http://localhost:8000/files/my-bucket?key=photos/2026/photo.jpg"
```

Успешный ответ:

```http
HTTP/1.1 200 OK
Content-Type: image/jpeg
```

Тело ответа стримится из S3 клиенту.

### Delete File

```http
DELETE /files/{bucket}?key={object-key}
```

Пример:

```bash
curl -i -X DELETE \
  "http://localhost:8000/files/my-bucket?key=photos/2026/photo.jpg"
```

Успешный ответ:

```http
HTTP/1.1 204 No Content
```

## ⚠️ Формат Ошибок

Ошибки возвращаются в JSON:

```json
{
  "err": "key is empty"
}
```

Основные статусы:

| Status | Meaning |
| --- | --- |
| `400 Bad Request` | Не передан `key`, некорректный multipart body или невалидный запрос |
| `404 Not Found` | Bucket или object не найден в S3 |
| `413 Payload Too Large` | Upload-запрос превысил лимит размера |
| `500 Internal Server Error` | Неожиданная ошибка storage или streaming |
| `503 Service Unavailable` | Readiness check не прошёл |

## ⚙️ Конфигурация

Приложение читает конфигурацию из переменных окружения. Локальный `.env` загружается, если он есть, но не является обязательным.

| Variable | Required | Description | Example |
| --- | --- | --- | --- |
| `HTTP_ADDR` | No | Адрес HTTP-сервера | `0.0.0.0:8000` |
| `S3_ENDPOINT` | Yes | S3-compatible endpoint | `http://s3gw-minio:9000` |
| `S3_REGION` | Yes | S3 region | `us-east-1` |
| `S3_KEY` | Yes | Access key | `minioadmin` |
| `S3_SECRET` | Yes | Secret key | `minioadmin` |

Пример `.env` для локального Docker Compose:

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

Файл `.env` передаётся в контейнер через Compose и исключён из Docker build context через `.dockerignore`.

## 🚀 Локальный Запуск

Запустить stack:

```bash
docker compose up --build
```

Сервисы:

| Service | URL |
| --- | --- |
| Gateway | `http://localhost:8000` |
| Gateway metrics | `http://localhost:8000/metrics` |
| MinIO S3 API | `http://localhost:9000` |
| MinIO Console | `http://localhost:9001` |
| Prometheus | `http://localhost:9090` |
| Grafana | `http://localhost:3000` |
| PostgreSQL | `localhost:5432` |

Перед загрузкой файлов нужно создать bucket в MinIO. Это можно сделать через MinIO Console или любой S3-compatible client.

## 📊 Observability

Gateway отдаёт Prometheus metrics на `/metrics`.

Compose stack включает:

- Prometheus, который scrape-ит `app:8000`;
- Grafana для dashboards;
- стандартные Go runtime и process metrics из коробки.

Полезные метрики, доступные сразу:

| Metric | Meaning |
| --- | --- |
| `go_goroutines` | Текущее количество goroutines |
| `go_memstats_heap_alloc_bytes` | Heap memory, выделенная процессом |
| `process_cpu_seconds_total` | CPU time, потреблённый процессом |
| `process_resident_memory_bytes` | Resident memory процесса |

## 🧰 Запуск Без Docker

Установить нужные переменные окружения и запустить:

```bash
go run ./cmd/main.go
```

Если `HTTP_ADDR` не задан, сервис использует:

```text
0.0.0.0:8000
```

## ✅ Тесты

Запустить все тесты:

```bash
go test ./...
```

Текущий набор тестов покрывает поведение HTTP handlers через fake storage:

- успешный upload;
- успешный download;
- успешный delete;
- mapping S3 not-found ошибок;
- закрытие response body после download.

Integration tests требуют запущенный S3-compatible backend.

Для запуска против локального MinIO:

```bash
docker compose up -d minio
```

Затем:

```bash
S3_ENDPOINT=http://localhost:9000 \
S3_REGION=us-east-1 \
S3_KEY=minioadmin \
S3_SECRET=minioadmin \
go test ./s3_storage
```

В PowerShell:

```powershell
$env:S3_ENDPOINT="http://localhost:9000"
$env:S3_REGION="us-east-1"
$env:S3_KEY="minioadmin"
$env:S3_SECRET="minioadmin"
go test ./s3_storage
```

Integration test создаёт временный bucket, загружает объект, скачивает его, удаляет и проверяет, что удалённый key больше недоступен.

## 🗂️ Структура Проекта

```text
cmd/
  main.go                   Точка входа и сборка зависимостей

config/
  config.go                 Конфигурация через environment variables

router/
  router.go                 Регистрация HTTP routes
  handlers/
    healthz.go              Health endpoint
    readyz.go               Readiness endpoint
    get_file.go             Download handler
    put_file.go             Upload handler
    delete_file.go          Delete handler
    helpers.go              Общие HTTP response helpers
    handlers_test.go        Unit tests для handlers

s3_storage/
  storage.go                S3-compatible storage adapter
  interface.go              Storage interface для handlers
  errors.go                 S3-specific error helpers
  storage_integration_test.go
                            MinIO/S3 integration test

Dockerfile                  Multi-stage image build
compose.yaml                Local stack: app, MinIO, PostgreSQL, Prometheus, Grafana
prometheus.yml              Prometheus scrape config
```

## 🧠 Design Notes

Gateway отделяет HTTP layer от storage adapter:

```text
HTTP client
  -> router handlers
  -> S3Storage interface
  -> AWS SDK / MinIO / S3-compatible storage
```

Handlers отвечают за HTTP-уровень: path/query parameters, multipart parsing, response codes и JSON errors.

Пакет `s3_storage` отвечает за S3-уровень: настройку AWS SDK client, path-style addressing, object streaming, content type handling и S3-specific error checks.

Так кодовая база остаётся небольшой, тестируемой и удобной для расширения без лишних слоёв.

## 🔐 Security And Operational Notes

- `.env` не копируется в Docker image.
- Размер upload body ограничен на HTTP-уровне.
- Файлы стримятся в storage и из storage без полной буферизации в памяти.
- Финальный Docker image запускается от non-root user.
- Логи пишутся в структурированном JSON-формате через `zap`.
- Docker image имеет healthcheck на `/healthz`.
- Prometheus и Grafana включены для локальной observability.

Potential next steps:

- authentication and authorization;
- request logging middleware;
- custom HTTP and S3 metrics;
- bucket creation workflow or bootstrap script;
- PostgreSQL-backed file metadata;
- more precise storage error mapping for upload failures.
