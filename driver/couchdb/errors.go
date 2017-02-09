package couchdb

import (
	"fmt"
	"net/http"
)

// HTTPError is an error that represents an HTTP transport error.
type HTTPError struct {
	StatusCode int
	Status     string
}

func (e *HTTPError) Error() string {
	return fmt.Sprintf("HTTP Error: %s", e.Status)
}

// ResponseError returns an error from the HTTP response.
func ResponseError(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	return &HTTPError{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
	}
}
