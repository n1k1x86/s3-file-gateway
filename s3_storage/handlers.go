package s3_storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func handleError(err error, w http.ResponseWriter, statusCode int) {
	w.Header().Add("Content-Type", "application/json")
	data, err := json.Marshal(&ErrResp{Err: err.Error()})
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(data)
}

type ErrResp struct {
	Err string `json:"err"`
}

func GetFile(s S3Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.URL.Query().Get("key")
		if key == "" {
			handleError(fmt.Errorf("key is empty"), w, http.StatusBadRequest)
			return
		}

		reader, contentType, err := s.GetObject(r.Context(), bucket, key)
		if err != nil {
			var noSuchBucketErr *types.NoSuchBucket
			var noFileErr *types.NoSuchKey

			if errors.As(err, &noSuchBucketErr) || errors.As(err, &noFileErr) {
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

func PutFile(s S3Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		file, header, err := r.FormFile("file")
		if err != nil {
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

		w.WriteHeader(http.StatusCreated)
	}
}

func DeleteFile(s S3Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		bucket := r.PathValue("bucket")
		key := r.URL.Query().Get("key")

		if key == "" {
			handleError(fmt.Errorf("key is empty"), w, http.StatusBadRequest)
			return
		}

		err := s.DeleteObject(r.Context(), bucket, key)
		if err != nil {
			var noSuchBucketErr *types.NoSuchBucket
			var noFileErr *types.NoSuchKey

			if errors.As(err, &noSuchBucketErr) || errors.As(err, &noFileErr) {
				handleError(err, w, http.StatusNotFound)
				return
			}
			handleError(err, w, http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
