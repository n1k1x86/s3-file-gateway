package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"s3-file-gateway/s3_storage"
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

type PutFileResp struct {
	Key string `json:"key"`
}

func PutFile(s s3_storage.S3Storage) http.HandlerFunc {
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

		respBody, err := json.Marshal(&PutFileResp{
			Key: key,
		})

		w.WriteHeader(http.StatusCreated)
		w.Write(respBody)
	}
}

func DeleteFile(s s3_storage.S3Storage) http.HandlerFunc {
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
