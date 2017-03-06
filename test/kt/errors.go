package kt

import (
	"net/http"
	"testing"

	"github.com/flimzy/kivik/errors"
)

// IsError checks a kivik error against the expected status code (0 = no
// error), and returns true if the expectation is unmet.
func IsError(err error, status int, t *testing.T) bool {
	switch errors.StatusCode(err) {
	case status:
		// This is expected
		return false
	case 0:
		t.Errorf("Expected failure %d/%s", status, http.StatusText(status))
		return true
	default:
		if status == 0 {
			t.Errorf("Unexpected failure: %s", err)
			return true
		}
		t.Errorf("Unexpected failure state.\nExpected: %d/%s\n  Actual: %s", status, http.StatusText(status), err)
		return true
	}
}
