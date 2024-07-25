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

// Package nettest wraps [httptest.NewServer] to skip when called from GopherJS.
package nettest

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// NewHTTPTestServer wraps [httptest.NewServer]
func NewHTTPTestServer(_ *testing.T, handler http.Handler) *httptest.Server {
	return httptest.NewServer(handler)
}
