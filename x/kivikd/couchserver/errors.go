package couchserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivikd/v4/internal"
)

func errorDescription(status int) string {
	switch status {
	case 401:
		return "unauthorized"
	case 400:
		return "bad_request"
	case 404:
		return "not_found"
	case 500:
		return "internal_server_error" // TODO: Validate that this is normative
	case 501:
		return "not_implemented" // Non-standard
	}
	panic(fmt.Sprintf("unknown status %d", status))
}

type couchError struct {
	Error  string `json:"error"`
	Reason string `json:"reason"`
}

// HandleError returns a CouchDB-formatted error. It does nothing if err is nil.
func (h *Handler) HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	status := kivik.HTTPStatus(err)
	w.WriteHeader(status)
	var reason string
	kerr := new(internal.Error)
	if errors.As(err, &kerr) {
		reason = kerr.Message
	} else {
		reason = err.Error()
	}
	wErr := json.NewEncoder(w).Encode(couchError{
		Error:  errorDescription(status),
		Reason: reason,
	})
	if wErr != nil {
		h.Logger.Printf("Failed to send send error: %s", wErr)
	}
}
