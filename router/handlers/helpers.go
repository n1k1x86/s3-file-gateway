package handlers

import (
	"encoding/json"
	"net/http"
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
