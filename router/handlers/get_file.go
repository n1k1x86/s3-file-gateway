package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"s3-file-gateway/s3_storage"

	"go.uber.org/zap"
)

func GetFile(s s3_storage.S3Storage, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.URL.Query().Get("key")
		if key == "" {
			logger.Error("key is empty")
			handleError(fmt.Errorf("key is empty"), w, http.StatusBadRequest)
			return
		}

		reader, contentType, err := s.GetObject(r.Context(), bucket, key)
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
			logger.Error("downloading file", zap.Error(err))
			handleError(err, w, http.StatusInternalServerError)
			return
		}

		defer reader.Close()

		w.Header().Add("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, reader)
		if err != nil {
			logger.Error("downloading file", zap.Error(err))
			handleError(err, w, http.StatusInternalServerError)
			return
		}
		logger.Info("downloaded file", zap.String("bucket", bucket), zap.String("key", key))
	}
}
