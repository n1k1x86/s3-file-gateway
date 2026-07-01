package router

import "net/http"

func GetFiles(w http.ResponseWriter, r *http.Request) {
	fileID := r.PathValue("id")

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("got file with id " + fileID))
}

func PostFiles(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("created file"))
}

func DeleteFiles(w http.ResponseWriter, r *http.Request) {
	_ = r.PathValue("id")

	w.WriteHeader(http.StatusNoContent)
}
