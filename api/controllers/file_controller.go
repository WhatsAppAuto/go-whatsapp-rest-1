package controllers

import (
	"net/http"

	"github.com/gorilla/mux"
)

func (server *Server) FileDownload(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	http.ServeFile(w, r, "./share/store/"+vars["id"])
	return
}
