package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

func newTestApplication(t *testing.T) *application {
	return &application{
		errorLog: log.New(io.Discard, "", 0),
		infoLog:  log.New(io.Discard, "", 0),
	}
}

type testServer struct {
	*httptest.Server
}

//Create a newTestServer helper which initializes and returns a new instance of custom testServer type
func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	// add the cookie jar to the test server client so that any response cookies will be stored and sent when using this client
	ts.Client().Jar = jar

	//disable the redirect-following for the test server by setting a custom CheckRedirect function
	ts.Client().CheckRedirect = func(r *http.Request, vua []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &testServer{ts}
}

//inplement a get() method on custom testServer, which makes a Get request to a given url using the 
//test server client, and returns the response statuscode, headers and the body
func (ts *testServer)get(t *testing.T, urlPath string)(int, http.Header,string){
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)

	if err != nil{
		t.Fatal(err)
	}
	bytes.TrimSpace(body)
	return rs.StatusCode,rs.Header, string(body)
}
