package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"s3-file-gateway/s3_storage"

	"go.uber.org/zap"
)

func DeleteFile(s s3_storage.S3Storage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.URL.Query().Get("key")

		if key == "" {
			logger.Error("key is empty")
			handleError(fmt.Errorf("key is empty"), w, http.StatusBadRequest)
			return
		}

		err := s.DeleteObject(r.Context(), bucket, key)
		if err != nil {
			if errors.As(err, &s3_storage.ErrNoSuchBucket) {
				logger.Error("not found buckets with such name", zap.String("bucket", bucket))
				handleError(err, w, http.StatusNotFound)
				return
			}

			if errors.As(err, &s3_storage.ErrNoSuchKey) {
				logger.Error("not found files with such key", zap.String("key", key))
				handleError(err, w, http.StatusNotFound)
				return
			}
			logger.Error("deleting object", zap.Error(err))
			handleError(err, w, http.StatusInternalServerError)
			return
		}

		logger.Info("deleted object", zap.String("bucket", bucket), zap.String("key", key))
		w.WriteHeader(http.StatusNoContent)
	}
}
