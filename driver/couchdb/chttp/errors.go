package chttp

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"

	"github.com/flimzy/kivik"
)

// HTTPError is an error that represents an HTTP transport error.
type HTTPError struct {
	Code   int
	Reason string `json:"reason"`
}

func (e *HTTPError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("%d %s", e.Code, http.StatusText(e.Code))
	}
	return fmt.Sprintf("%d %s: %s", e.Code, http.StatusText(e.Code), e.Reason)
}

// StatusCode returns the embedded status code.
func (e *HTTPError) StatusCode() int {
	return e.Code
}

// StatusCode returns the status code of the error.
func StatusCode(err error) int {
	if err == nil {
		return 0
	}
	if httperr, ok := err.(*HTTPError); ok {
		return httperr.Code
	}
	return http.StatusInternalServerError
}

// ResponseError returns an error from an *http.Response.
func ResponseError(resp *http.Response) error {
	if resp.StatusCode < 400 {
		return nil
	}
	httpErr := &HTTPError{}
	if resp.Request.Method != "HEAD" && resp.ContentLength != 0 {
		if ct, _, _ := mime.ParseMediaType(resp.Header.Get("Content-Type")); ct == kivik.TypeJSON {
			dec := json.NewDecoder(resp.Body)
			defer resp.Body.Close()
			if err := dec.Decode(httpErr); err != nil {
				fmt.Printf("Failed to decode error response: %s\n", err)
			}
		}
	}
	httpErr.Code = resp.StatusCode
	return httpErr
}
