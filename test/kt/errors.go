package kt

import (
	"net/http"

	"github.com/flimzy/kivik/errors"
)

// CheckError compares the error's status code with that expected.
func (c *Context) CheckError(err error) (match bool, success bool) {
	status := c.Int("status")
	if status == 0 && err == nil {
		return true, true
	}
	switch errors.StatusCode(err) {
	case status:
		// This is expected
		return true, status == 0
	case 0:
		c.Errorf("Expected failure %d/%s", status, http.StatusText(status))
		return false, true
	default:
		if status == 0 {
			c.Errorf("Unexpected failure: %d/%s", errors.StatusCode(err), err)
			return false, false
		}
		c.Errorf("Unexpected failure state.\nExpected: %d/%s\n  Actual: %d/%s", status, http.StatusText(status), errors.StatusCode(err), err)
		return false, false
	}
}

// IsExpected checks the error against the expected status, and returns true
// if they match.
func (c *Context) IsExpected(err error) bool {
	m, _ := c.CheckError(err)
	return m
}

// IsSuccess is similar to IsExpected, except for its return value. This method
// returns true if the expected status == 0, regardless of the error.
func (c *Context) IsSuccess(err error) bool {
	_, s := c.CheckError(err)
	return s
}

// IsExpectedSuccess combines IsExpected() and IsSuccess(), returning true only
// if there is no error, and no error was expected.
func (c *Context) IsExpectedSuccess(err error) bool {
	m, s := c.CheckError(err)
	return m && s
}
