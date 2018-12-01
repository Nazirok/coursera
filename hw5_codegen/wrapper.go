package main

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
)

type resp map[string]interface{}

func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusNotAcceptable)
			data, _ := json.Marshal(resp{"error": "bad method"})
			w.Write(data)
			return
		}
		srv.handlerPOST(w, r)
	case "/user/profile":
		if !(r.Method == http.MethodGet || r.Method == http.MethodPost) {
			w.WriteHeader(http.StatusNotAcceptable)
			data, _ := json.Marshal(resp{"error": "bad method"})
			w.Write(data)
			return
		}
		srv.handlerGET(w, r)
	default:
		w.WriteHeader(http.StatusNotFound)
		data, _ := json.Marshal(resp{"error": "unknown method"})
		w.Write(data)
		return
	}
}

func (srv *MyApi) handlerPOST(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if at := r.Header.Get("X-Auth"); at != "100500" {
		w.WriteHeader(http.StatusForbidden)
		data, _ := json.Marshal(resp{"error": "unauthorized"})
		w.Write(data)
		return
	}

	login := r.FormValue("login")
	if login == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "login must me not empty"})
		w.Write(data)
		return
	}
	if len(login) < 10 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "login len must be >= 10"})
		w.Write(data)
		return
	}

	full_name := r.FormValue("full_name")

	status := r.FormValue("status")
	if status == "" {
		status = "user"
	} else if !(status == "user" || status == "moderator" || status == "admin") {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "status must be one of [user, moderator, admin]"})
		w.Write(data)
		return
	}

	a := r.FormValue("age")
	age, err := strconv.Atoi(a)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "age must be int"})
		w.Write(data)
		return
	}
	if age < 0 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "age must be >= 0"})
		w.Write(data)
		return
	}
	if age > 128 {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "age must be <= 128"})
		w.Write(data)
		return
	}
	in := CreateParams{
		Login:  login,
		Name:   full_name,
		Status: status,
		Age:    age,
	}
	user, err := srv.Create(context.Background(), in)
	if err != nil {
		if v, ok := err.(ApiError); ok {
			w.WriteHeader(v.HTTPStatus)
			data, _ := json.Marshal(resp{"error": v.Error()})
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		data, _ := json.Marshal(resp{"error": err.Error()})
		w.Write(data)
		return
	}

	response := map[string]interface{}{
		"error":    "",
		"response": user,
	}
	data, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return
}

func (srv *MyApi) handlerGET(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	login := r.FormValue("login")
	if login == "" {
		w.WriteHeader(http.StatusBadRequest)
		data, _ := json.Marshal(resp{"error": "login must me not empty"})
		w.Write(data)
		return
	}
	in := ProfileParams{
		Login: login,
	}
	user, err := srv.Profile(context.Background(), in)
	if err != nil {
		if v, ok := err.(ApiError); ok {
			w.WriteHeader(v.HTTPStatus)
			data, _ := json.Marshal(resp{"error": v.Error()})
			w.Write(data)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		data, _ := json.Marshal(resp{"error": err.Error()})
		w.Write(data)
		return
	}
	response := map[string]interface{}{
		"error":    "",
		"response": user,
	}
	data, _ := json.Marshal(response)
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	return

}
