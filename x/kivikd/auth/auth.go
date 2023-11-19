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
	"encoding/json"
	"net/http"

	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
)

// Handler is an auth handler.
type Handler interface {
	// MethodName identifies the handler. It is only called once on server
	// start up.
	MethodName() string
	// Authenticate authenticates the HTTP request. On success, a user context
	// must be returned. Any error will immediately terminate the authentication
	// process, returning an error to the client. In particular, this means that
	// an "unauthorized" error must not be returned if fallthrough is intended.
	// If a response is sent, execution does not continue. This allows handlers
	// to expose their own API endpoints (for example, the default cookie auth
	// handler adds POST /_session and DELETE /_session handlers).
	Authenticate(http.ResponseWriter, *http.Request) (*authdb.UserContext, error)
}

// Session represents an authenticated session.
type Session struct {
	AuthMethod string
	AuthDB     string
	Handlers   []string
	User       *authdb.UserContext
}

// MarshalJSON satisfies the json.Marshaler interface.
func (s *Session) MarshalJSON() ([]byte, error) {
	user := s.User
	if user == nil {
		user = &authdb.UserContext{}
	}
	result := map[string]interface{}{
		"info": map[string]interface{}{
			"authenticated":           s.AuthMethod,
			"authentication_db":       s.AuthDB,
			"authentication_handlers": s.Handlers,
		},
		"ok":      true,
		"userCtx": user,
	}
	return json.Marshal(result)
}
