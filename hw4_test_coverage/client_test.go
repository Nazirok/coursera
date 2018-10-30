package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// код писать тут

const accessToken = "someToken"

func SearchServer(w http.ResponseWriter, r *http.Request) {
	at := r.Header.Get("AccessToken")
	if at != accessToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
}

func TestUnauthorized(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "some",
		URL:         ts.URL,
	}
	searcherReq := SearchRequest{}
	_, err := client.FindUsers(searcherReq)
	if err.Error() != "Bad AccessToken" {
		t.Error("Test Unauthorized Failed")
	}
}

func TestRequestTimeout(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2500 * time.Millisecond)
	}))

	client := SearchClient{URL: ts.URL}
	searcherReq := SearchRequest{}
	_, err := client.FindUsers(searcherReq)
	if !strings.Contains(err.Error(), "timeout for") {
		t.Error("Test timeout failed")
	}
}

func TestUnknownError(t *testing.T) {
	client := SearchClient{}
	searcherReq := SearchRequest{}
	_, err := client.FindUsers(searcherReq)
	if !strings.Contains(err.Error(), "unknown error") {
		t.Error("Test unknown error failed")
	}
}

func TestBadLimitAndOffset(t *testing.T) {
	client := SearchClient{}
	searcherReqLimit := SearchRequest{
		Limit: -1,
	}
	_, err := client.FindUsers(searcherReqLimit)
	if !strings.Contains(err.Error(), "limit must be > 0") {
		t.Error("Test bad limit failed")
	}

	searcherReqOffset := SearchRequest{
		Limit:  26,
		Offset: -1,
	}
	_, err = client.FindUsers(searcherReqOffset)
	if !strings.Contains(err.Error(), "offset must be > 0") {
		t.Error("Test bad offset failed")
	}
}

func TestInternalServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	client := SearchClient{URL: ts.URL}
	searcherReq := SearchRequest{}
	_, err := client.FindUsers(searcherReq)
	if !strings.Contains(err.Error(), "SearchServer fatal error") {
		t.Error("Test internal server error failed")
	}
}

