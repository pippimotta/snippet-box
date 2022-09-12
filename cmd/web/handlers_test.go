package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pippimotta/snippet-box/internal/assert"
)

func TestPing(t *testing.T) {
	app := &application{
		errorLog: log.New(io.Discard,"",0),
		infoLog : log.New(io.Discard,"",0),
	}

	ts := httptest.NewTLSServer(app.routes())
	defer ts.Close()

	//use ts.Client() to mock a client and use Get() to make a request to the test server
	rs, err := ts.Client().Get(ts.URL +"/ping")

	if err != nil {
		t.Fatal(err)
	}

	//then check the value of the response status code and body using the same pattern
	assert.Equal(t, rs.StatusCode, http.StatusOK)
	defer rs.Body.Close()

	//check if the response body written by the ping handler equals OK
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)
	assert.Equal(t, string(body), "OK")

}
