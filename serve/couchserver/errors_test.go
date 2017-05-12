package couchserver

import (
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/flimzy/diff"
)

type reasonError int

func (e reasonError) Error() string   { return "test error" }
func (e reasonError) StatusCode() int { return 404 }
func (e reasonError) Reason() string  { return "it ain't there" }

func TestHandleError(t *testing.T) {
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
				HandleError(w, test.Err)
				resp := w.Result()
				defer resp.Body.Close()
				if d := diff.AsJSON(test.Expected, resp.Body); d != "" {
					t.Error(d)
				}
			})
		}(test)
	}
}
