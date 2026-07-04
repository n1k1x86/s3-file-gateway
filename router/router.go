package router

import (
	"s3-file-gateway/router/handlers"
	"s3-file-gateway/s3_storage"

	"github.com/n1k1x86/libs/http_server"
	"go.uber.org/zap"
)

func InitRouter(storage s3_storage.S3Storage, logger *zap.Logger) http_server.HTTPMux {
	mux := http_server.NewMux()

	mux.HandleFunc("GET /files/{bucket}", handlers.GetFile(storage, logger))
	mux.HandleFunc("POST /files/{bucket}", handlers.PutFile(storage, logger))
	mux.HandleFunc("DELETE /files/{bucket}", handlers.DeleteFile(storage, logger))

	mux.HandleFunc("GET /healthz", Healthz)

	return mux
}
