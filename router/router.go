package router

import "github.com/n1k1x86/libs/http_server"

func InitRouter() http_server.HTTPMux {
	mux := http_server.NewMux()

	mux.HandleFunc("GET /files/{id}", GetFiles)
	mux.HandleFunc("POST /files", PostFiles)
	mux.HandleFunc("DELETE /files/{id}", DeleteFiles)

	return mux
}
