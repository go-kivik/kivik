package couchserver

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kivik/kivikd/v4/internal"
	"gitlab.com/flimzy/testy"
)

func TestHandleError(t *testing.T) {
	h := Handler{}
	type eTest struct {
		Name     string
		Err      error
		Expected interface{}
	}
	tests := []eTest{
		{
			Name:     "NilError",
			Expected: nil,
		},
		{
			Name: "SimpleError",
			Err:  errors.New("test error"),
			Expected: map[string]string{
				"error":  "internal_server_error",
				"reason": "test error",
			},
		},
		{
			Name: "kivik error",
			Err:  &internal.Error{Status: http.StatusNotFound, Message: "it ain't there"},
			Expected: map[string]string{
				"error":  "not_found",
				"reason": "it ain't there",
			},
		},
	}
	for _, test := range tests {
		func(test eTest) {
			t.Run(test.Name, func(t *testing.T) {
				w := httptest.NewRecorder()
				h.HandleError(w, test.Err)
				resp := w.Result()
				defer resp.Body.Close()
				if d := testy.DiffAsJSON(test.Expected, resp.Body); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

type errorResponseWriter struct {
	http.ResponseWriter
}

func (w *errorResponseWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("unusual write error")
}

func TestHandleErrorFailure(t *testing.T) {
	logBuf := &bytes.Buffer{}
	h := &Handler{
		Logger: log.New(logBuf, "", 0),
	}
	w := httptest.NewRecorder()
	h.HandleError(&errorResponseWriter{w}, errors.New("test error"))

	expected := "Failed to send send error: unusual write error\n"
	if expected != logBuf.String() {
		t.Errorf("Expected: %s\n  Actual: %s", expected, logBuf.String())
	}
}
