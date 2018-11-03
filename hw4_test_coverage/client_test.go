package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
)

// код писать тут

const (
	accessToken        = "someToken"
	xmlPath     string = "dataset.xml"
)

type UserXml struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	at := r.Header.Get("AccessToken")
	if at != accessToken {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	query := r.FormValue("query")
	orderField := r.FormValue("order_field")
	limit := r.FormValue("limit")
	//offset := r.FormValue("offset")
	orderBy := r.FormValue("order_by")

	if !(orderField == "Id" || orderField == "Age" || orderField == "Name") {
		w.WriteHeader(http.StatusBadRequest)
		errMsg := struct {
			Error string
		}{Error: "ErrorBadOrderField"}
		out, _ := json.Marshal(errMsg)
		w.Write(out)
		return
	}
	if !(orderBy == "-1" || orderBy == "0" || orderBy == "1") {
		w.WriteHeader(http.StatusBadRequest)
		errMsg := struct {
			Error string
		}{Error: "BadOrderByField"}
		out, _ := json.Marshal(errMsg)
		w.Write(out)
		return
	}
	usersXml := getXml(query, limit)
	l := len(usersXml)
	out := make([]User, 0, l)
	for i := 0; i < l; i++ {
		out = append(out, User{
			Name:   usersXml[i].FirstName + usersXml[i].LastName,
			Age:    usersXml[i].Age,
			About:  usersXml[i].About,
			Gender: usersXml[i].Gender,
			Id:     usersXml[i].ID,
		})
	}
	js, _ := json.Marshal(out)
	w.Write(js)

}

func getXml(query, limit string) []UserXml {
	file, err := os.Open(xmlPath)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)
	users := make([]UserXml, 0)
	for {
		tok, tokenErr := decoder.Token()
		if tokenErr != nil && tokenErr != io.EOF {
			fmt.Println("error happend", tokenErr)
			break
		} else if tokenErr == io.EOF {
			break
		}
		if tok == nil {
			fmt.Println("t is nil break")
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			if tok.Name.Local == "row" {
				user := UserXml{}
				if err := decoder.DecodeElement(&user, &tok); err != nil {
					fmt.Println("error happend", err)
				}
				if strings.Contains(user.FirstName+user.LastName, query) || strings.Contains(user.About, query) {
					users = append(users, user)
				}
			}
		}
	}
	out, _ := strconv.Atoi(limit)
	return users[0:out]
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

func TestBadOrderFiled(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "someToken",
		URL:         ts.URL,
	}
	searcherReq := SearchRequest{
		OrderField: "someFiled",
	}
	_, err := client.FindUsers(searcherReq)
	if err.Error() != fmt.Sprintf("OrderFeld %s invalid", searcherReq.OrderField) {
		t.Error("Test bad order filed Failed")
	}
}

func TestJsonUnpackError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := []byte{10, 55, 67, 34}
		w.Write(b)
	}))

	client := SearchClient{URL: ts.URL}
	searcherReq := SearchRequest{}
	_, err := client.FindUsers(searcherReq)
	if !strings.Contains(err.Error(), "cant unpack result json") {
		t.Error("Test unpack json filed failed")
	}
}

func TestJsonUnpackErrorBadRequest(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := []byte{10, 55, 67, 34}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(b)
	}))

	client := SearchClient{URL: ts.URL}
	searcherReq := SearchRequest{}
	_, err := client.FindUsers(searcherReq)
	if !strings.Contains(err.Error(), "cant unpack error json") {
		t.Error("Test unpack json bad request failed")
	}
}

func TestUnknownResponseError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "someToken",
		URL:         ts.URL,
	}
	searcherReq := SearchRequest{
		OrderField: "Age",
		OrderBy:    6,
	}
	_, err := client.FindUsers(searcherReq)
	if !strings.Contains(err.Error(), "unknown bad request") {
		t.Error("Test unknown response error failed")
	}
}

func TestGetResultsWithLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(SearchServer))

	client := SearchClient{
		AccessToken: "someToken",
		URL:         ts.URL,
	}
	searcherReq := SearchRequest{
		OrderField: "Age",
		Limit:      10,
	}
	resp, err := client.FindUsers(searcherReq)
	if err != nil {
		t.Error("test get results failed")
	}
	if len(resp.Users) != 10 {
		t.Error("test get results failed")
	}
}

func TestGetResultsWithSmallLimit(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u := User{
			Name:   "AAAAAAAAAAAA",
			Age:    22,
			Gender: "Male",
			About:  "xcvxcvxcv",
			Id:     4,
		}
		out := []User{u}
		ls, _ := json.Marshal(out)
		w.Write(ls)
	}))

	client := SearchClient{
		AccessToken: "someToken",
		URL:         ts.URL,
	}
	searcherReq := SearchRequest{
		OrderField: "Age",
		Limit:      10,
	}

	resp, err := client.FindUsers(searcherReq)
	if err != nil {
		t.Error("test get results failed")
	}
	if len(resp.Users) != 1 {
		t.Error("test get results with small limit failed")
	}
}
