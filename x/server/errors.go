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

package server

import "net/http"

var errNotImplimented = &couchError{status: http.StatusNotImplemented, Err: "not_implemented", Reason: "Feature not implemented"}

type couchError struct {
	status int
	Err    string `json:"error"`
	Reason string `json:"reason"`
}

func (e *couchError) Error() string {
	return e.Reason
}

func (e *couchError) HTTPStatus() int {
	return e.status
}
