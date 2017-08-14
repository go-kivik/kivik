package couchserver

import (
	"bytes"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/flimzy/diff"
)

type reasonError int

func (e reasonError) Error() string   { return "test error" }
func (e reasonError) StatusCode() int { return 404 }
func (e reasonError) Reason() string  { return "it ain't there" }

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
			Name: "ReasonError",
			Err:  reasonError(0),
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
				if d := diff.AsJSON(test.Expected, resp.Body); d != nil {
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

type fileResponseWriter struct {
	f *os.File
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
