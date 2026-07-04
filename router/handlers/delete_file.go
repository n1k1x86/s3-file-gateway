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
			handleError(fmt.Errorf("key is empty"), w, http.StatusBadRequest)
			return
		}

		err := s.DeleteObject(r.Context(), bucket, key)
		if err != nil {
			if errors.As(err, &s3_storage.ErrNoSuchBucket) || errors.As(err, &s3_storage.ErrNoSuchKey) {
				handleError(err, w, http.StatusNotFound)
				return
			}
			handleError(err, w, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
