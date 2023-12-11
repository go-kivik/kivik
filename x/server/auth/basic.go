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

package auth

import (
	"net/http"
)

type basicAuth struct {
	s Server
}

// BasicAuth returns a basic auth handler.
func BasicAuth() Handler {
	return &basicAuth{}
}

func (a *basicAuth) Init(s Server) (string, AuthenticateFunc) {
	a.s = s
	return "default", // For compatibility with the name used by CouchDB
		a.Authenticate
}

func (a *basicAuth) Authenticate(_ http.ResponseWriter, r *http.Request) (*UserContext, error) {
	username, password, ok := r.BasicAuth()
	if !ok {
		return nil, nil
	}
	return a.s.UserStore().Validate(r.Context(), username, password)
}
