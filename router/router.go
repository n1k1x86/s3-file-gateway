package router

import (
	"s3-file-gateway/s3_storage"

	"github.com/n1k1x86/libs/http_server"
)

func InitRouter(storage s3_storage.S3Storage) http_server.HTTPMux {
	mux := http_server.NewMux()

	mux.HandleFunc("GET /files/{bucket}", GetFile(storage))
	mux.HandleFunc("POST /files/{bucket}", PutFile(storage))
	mux.HandleFunc("DELETE /files/{bucket}", DeleteFile(storage))

	mux.HandleFunc("GET /healthz", Healthz)

	return mux
}
