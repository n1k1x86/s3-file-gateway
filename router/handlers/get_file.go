package handlers

import (
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
			logger.Warn("key is empty")
			handleError(fmt.Errorf("key is empty"), w, http.StatusBadRequest)
			return
		}

		reader, contentType, err := s.GetObject(r.Context(), bucket, key)
		if err != nil {
			if s3_storage.IsErrNoSuchBucket(err) {
				logger.Warn("not found buckets with such name", zap.String("bucket", bucket))
				handleError(err, w, http.StatusNotFound)
				return
			}

			if s3_storage.IsErrNoSuchKey(err) {
				logger.Warn("not found files with such key", zap.String("key", key))
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
