package handlers

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"s3-file-gateway/s3_storage"
)

func GetFile(s s3_storage.S3Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.URL.Query().Get("key")
		if key == "" {
			handleError(fmt.Errorf("key is empty"), w, http.StatusBadRequest)
			return
		}

		reader, contentType, err := s.GetObject(r.Context(), bucket, key)
		if err != nil {
			if errors.As(err, &s3_storage.ErrNoSuchBucket) || errors.As(err, &s3_storage.ErrNoSuchKey) {
				handleError(err, w, http.StatusNotFound)
				return
			}
			handleError(err, w, http.StatusInternalServerError)
			return
		}

		defer reader.Close()

		w.Header().Add("Content-Type", contentType)
		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, reader)
		if err != nil {
			handleError(err, w, http.StatusInternalServerError)
			return
		}
	}
}
