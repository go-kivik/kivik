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

package bindings

import (
	"testing"

	"github.com/gopherjs/gopherjs/js"

	_ "github.com/go-kivik/kivik/v4/pouchdb/bindings/poucherr"
)

type statuser interface { // nolint:misspell
	HTTPStatus() int
}

func TestNewPouchError(t *testing.T) {
	type npeTest struct {
		Name           string
		Object         *js.Object
		ExpectedStatus int
		Expected       string
	}
	tests := []npeTest{
		{
			Name:     "Null",
			Object:   nil,
			Expected: "",
		},
		{
			Name: "NameAndReasonNoStatus",
			Object: func() *js.Object {
				o := js.Global.Get("Object").New()
				o.Set("reason", "error reason")
				o.Set("name", "error name")
				return o
			}(),
			ExpectedStatus: 500,
			Expected:       "error name: error reason",
		},
		{
			Name: "ECONNREFUSED",
			Object: js.Global.Call("ReconstitutePouchError", `{
                "code":    "ECONNREFUSED",
                "errno":   "ECONNREFUSED",
                "syscall": "connect",
                "address": "127.0.0.1",
                "port":    5984,
                "status":  500,
                "result": {
                        "ok": false,
                        "start_time": "Tue May 16 2017 08:26:31 GMT+0000 (UTC)",
                        "docs_read": 0,
                        "docs_written": 0,
                        "doc_write_failures": 0,
                        "errors": [],
                        "status": "aborting",
                        "end_time": "Tue May 16 2017 08:26:31 GMT+0000 (UTC)",
                        "last_seq": 0
                    }
                }`),
			ExpectedStatus: 500,
			Expected:       "Error: connection refused",
		},
	}
	for _, test := range tests {
		func(test npeTest) {
			t.Run(test.Name, func(t *testing.T) {
				result := NewPouchError(test.Object)
				var msg string
				if result != nil {
					msg = result.Error()
				}
				if msg != test.Expected {
					t.Errorf("Expected error: %s\n  Actual error: %s", test.Expected, msg)
				}
				if result == nil {
					return
				}
				status := result.(statuser).HTTPStatus() // nolint:misspell
				if status != test.ExpectedStatus {
					t.Errorf("Expected status %d, got %d", test.ExpectedStatus, status)
				}
			})
		}(test)
	}
}
