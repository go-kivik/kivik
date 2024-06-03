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

//go:build js
// +build js

package bindings

import (
	"fmt"
	"net/http"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"

	internal "github.com/go-kivik/kivik/v4/int/errors"
)

type pouchError struct {
	*js.Object
	Err     string
	Message string
	Status  int
}

// NewPouchError parses a PouchDB error.
func NewPouchError(o *js.Object) error {
	if o == nil || o == js.Undefined {
		return nil
	}
	status := o.Get("status").Int()
	if status == 0 {
		status = http.StatusInternalServerError
	}

	var err, msg string
	switch {
	case o.Get("reason") != js.Undefined:
		msg = o.Get("reason").String()
	case o.Get("message") != js.Undefined:
		msg = o.Get("message").String()
	default:
		if jsbuiltin.InstanceOf(o, js.Global.Get("Error")) {
			return &internal.Error{Status: status, Message: o.Get("message").String()}
		}
	}
	switch {
	case o.Get("name") != js.Undefined:
		err = o.Get("name").String()
	case o.Get("error") != js.Undefined:
		err = o.Get("error").String()
	}

	if msg == "" && o.Get("errno") != js.Undefined {
		switch o.Get("errno").String() {
		case "ECONNREFUSED":
			msg = "connection refused"
		case "ECONNRESET":
			msg = "connection reset by peer"
		case "EPIPE":
			msg = "broken pipe"
		case "ETIMEDOUT", "ESOCKETTIMEDOUT":
			msg = "operation timed out"
		}
	}

	return &pouchError{
		Err:     err,
		Message: msg,
		Status:  status,
	}
}

func (e *pouchError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Message)
}

func (e *pouchError) HTTPStatus() int {
	return e.Status
}
