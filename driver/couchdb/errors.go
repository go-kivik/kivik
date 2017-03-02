package couchdb

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
)

// HTTPError is an error that represents an HTTP transport error.
type HTTPError struct {
	Code   int
	Status string
	Reason string `json:"reason"`
}

func (e *HTTPError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("HTTP Error: %s", e.Status)
	}
	return fmt.Sprintf("HTTP Error: %s: %s", e.Status, e.Reason)
}

// StatusCode returns the embedded status code.
func (e *HTTPError) StatusCode() int {
	return e.Code
}

// StatusCode returns the status code of the error.
func StatusCode(err error) int {
	if httperr, ok := err.(*HTTPError); ok {
		return httperr.Code
	}
	return 0
}

// ResponseError returns an error from the HTTP response.
func ResponseError(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	httpErr := &HTTPError{}
	if resp.Request.Method != "HEAD" && resp.ContentLength > 0 {
		if cType, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type")); cType == typeJSON {
			dec := json.NewDecoder(resp.Body)
			defer resp.Body.Close()
			if err := dec.Decode(httpErr); err != nil {
				fmt.Printf("Failed to decode error response: %s\n", err)
			}
		}
	}
	httpErr.Code = resp.StatusCode
	httpErr.Status = resp.Status
	return httpErr
}
