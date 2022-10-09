package main

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alexedwards/scs/v2"
	"github.com/go-playground/form/v4"
	"github.com/pippimotta/snippet-box/internal/models/mocks"
)

func newTestApplication(t *testing.T) *application {

	templateCache, err := newTemplateCache()
	if err != nil {
		t.Fatal(err)
	}

	formDecoder := form.NewDecoder()
	sessionManager := scs.New()
	sessionManager.Lifetime = 12 * time.Hour
	sessionManager.Cookie.Secure = true

	return &application{
		errorLog:       log.New(io.Discard, "", 0),
		infoLog:        log.New(io.Discard, "", 0),
		snippets:       &mocks.SnippetModel{},
		users:          &mocks.UserModel{},
		templateCache:  templateCache,
		formDecoder:    formDecoder,
		sessionManager: sessionManager,
	}
}

type testServer struct {
	*httptest.Server
}

// Create a newTestServer helper which initializes and returns a new instance of custom testServer type
func newTestServer(t *testing.T, h http.Handler) *testServer {
	ts := httptest.NewTLSServer(h)

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal(err)
	}
	// add the cookie jar to the test server client so that any response cookies will be stored and sent when using this client
	ts.Client().Jar = jar

	//disable the redirect-following for the test server by setting a custom CheckRedirect function
	ts.Client().CheckRedirect = func(r *http.Request, via []*http.Request) error {
		return http.ErrUseLastResponse
	}
	return &testServer{ts}
}

// inplement a get() method on custom testServer, which makes a Get request to a given url using the
// test server client, and returns the response statuscode, headers and the body
func (ts *testServer) get(t *testing.T, urlPath string) (int, http.Header, string) {
	rs, err := ts.Client().Get(ts.URL + urlPath)
	if err != nil {
		t.Fatal(err)
	}
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)

	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)

	return rs.StatusCode, rs.Header, string(body)
}
