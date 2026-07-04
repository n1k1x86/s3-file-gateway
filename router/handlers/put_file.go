package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"s3-file-gateway/s3_storage"

	"go.uber.org/zap"
)

var ErrHttpMaxBytes *http.MaxBytesError

const (
	maxUploadSize  = 50 << 20 // 50 MB
	maxRequestSize = maxUploadSize + 1<<20
)

type PutFileResp struct {
	Key string `json:"key"`
}

func PutFile(s s3_storage.S3Storage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, maxRequestSize)

		file, header, err := r.FormFile("file")
		if err != nil {
			if errors.As(err, &ErrHttpMaxBytes) {
				handleError(err, w, http.StatusRequestEntityTooLarge)
			}
			handleError(err, w, http.StatusBadRequest)
			return
		}
		defer file.Close()

		bucket := r.PathValue("bucket")
		key := r.URL.Query().Get("key")

		err = s.PutObject(r.Context(), bucket, key, file, header.Header.Get("Content-Type"))
		if err != nil {
			handleError(err, w, http.StatusBadRequest)
			return
		}

		respBody, err := json.Marshal(&PutFileResp{
			Key: key,
		})

		w.WriteHeader(http.StatusCreated)
		w.Write(respBody)
	}
}
