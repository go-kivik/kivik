// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:build !js
// +build !js

package couchserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func errorDescription(status int) string {
	switch status {
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusInternalServerError:
		return "internal_server_error" // TODO: Validate that this is normative
	case http.StatusNotImplemented:
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
