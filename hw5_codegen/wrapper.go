package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type respError struct {
	Error string `json:"error"`
}

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodPost:
		srv.handlerPOST(w, r)
	case http.MethodGet:
		srv.handlerGET(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func (srv *MyApi) handlerPOST(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	switch r.URL.Path {
	case "/user/create":
		if at := r.Header.Get("X-Auth"); at != "100500" {
			w.WriteHeader(http.StatusForbidden)
			data, _ := json.Marshal(respError{Error: "unauthorized"})
			w.Write(data)
			return
		}
		login := r.FormValue("login")
		if login == "" {
			w.WriteHeader(http.StatusBadRequest)
			data, _ := json.Marshal(respError{Error: "login must me not empty"})
			w.Write(data)
			return
		}
		if len(login) < 10 {
			w.WriteHeader(http.StatusBadRequest)
			data, _ := json.Marshal(respError{Error: "login len must be >= 10"})
			w.Write(data)
			return
		}
	}
	fmt.Println(r.URL.Path)
}

func (srv *MyApi) handlerGET(w http.ResponseWriter, r *http.Request) {

}