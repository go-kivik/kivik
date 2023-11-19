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

// Package basic provides HTTP Basic Auth services.
package basic

import (
	"net/http"

	"github.com/go-kivik/kivik/v4/x/kivikd"
	"github.com/go-kivik/kivik/v4/x/kivikd/auth"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
)

// HTTPBasicAuth provides HTTP Basic Auth
type HTTPBasicAuth struct{}

var _ auth.Handler = &HTTPBasicAuth{}

// MethodName returns "default"
func (a *HTTPBasicAuth) MethodName() string {
	return "default" // For compatibility with the name used by CouchDB
}

// Authenticate authenticates a request against a user store using HTTP Basic
// Auth.
func (a *HTTPBasicAuth) Authenticate(_ http.ResponseWriter, r *http.Request) (*authdb.UserContext, error) {
	store := kivikd.GetService(r).UserStore
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, nil
	}
	return store.Validate(r.Context(), username, password)
}
