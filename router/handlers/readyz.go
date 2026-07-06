package handlers

import (
	"net/http"
	"s3-file-gateway/s3_storage"

	"go.uber.org/zap"
)

const (
	readyzBucket = "my-bucket"
)

func Readyz(s3Storage s3_storage.S3Storage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := s3Storage.IsReady(r.Context(), readyzBucket)
		if err != nil {
			logger.Error("s3 is not ready", zap.String("bucket", readyzBucket))
			handleError(err, w, http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}
}
