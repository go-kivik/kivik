package couchdb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// HTTPError is an error that represents an HTTP transport error.
type HTTPError struct {
	StatusCode int
	Status     string
	Reason     string `json:"reason"`
}

func (e *HTTPError) Error() string {
	if e.Reason == "" {
		return fmt.Sprintf("HTTP Error: %s", e.Status)
	}
	return fmt.Sprintf("HTTP Error: %s: %s", e.Status, e.Reason)
}

// ResponseError returns an error from the HTTP response.
func ResponseError(resp *http.Response) error {
	if resp.StatusCode < 300 {
		return nil
	}
	err := &HTTPError{}
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(err); err != nil {
		fmt.Printf("Failed to decode error response: %s", err)
	}
	err.StatusCode = resp.StatusCode
	err.Status = resp.Status
	return err
}
