package kivikd

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/go-kivik/kivikd/v4/auth"
	"github.com/go-kivik/kivikd/v4/authdb"
	"github.com/go-kivik/kivikd/v4/logger"
)

func TestLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.New(buf)
	mw := loggerMiddleware(l)
	req := httptest.NewRequest("GET", "/foo", nil)
	session := &auth.Session{User: &authdb.UserContext{Name: "bob"}}
	ctx := context.WithValue(req.Context(), SessionKey, &session)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Go away!"))
	})
	mw(handler).ServeHTTP(w, req)
	expectedRE := regexp.MustCompile(`^192\.0\.2\.1 bob \[\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.+] ` +
		`\([\d.]+[µn]s\) "GET /foo HTTP/1.1" 401 8 "" ""`)
	if !expectedRE.Match(buf.Bytes()) {
		t.Errorf("Log does not match. Got:\n%s\n", buf.String())
	}
}

func TestLoggerNoAuth(t *testing.T) {
	buf := &bytes.Buffer{}
	l := logger.New(buf)
	mw := loggerMiddleware(l)
	req := httptest.NewRequest("GET", "/foo", nil)
	session := &auth.Session{}
	ctx := context.WithValue(req.Context(), SessionKey, &session)
	req = req.WithContext(ctx)
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("Go away!"))
	})
	mw(handler).ServeHTTP(w, req)
	expectedRE := regexp.MustCompile(`^192\.0\.2\.1  \[\d\d\d\d-\d\d-\d\d \d\d:\d\d:\d\d.+] ` +
		`\([\d.]+[µn]s\) "GET /foo HTTP/1.1" 401 8 "" ""`)
	if !expectedRE.Match(buf.Bytes()) {
		t.Errorf("Log does not match. Got:\n%s\n", buf.String())
	}
}
