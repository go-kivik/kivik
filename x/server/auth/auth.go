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

// Package auth provides authentication and authorization for the server.
package auth

import (
	"net/http"
)

// CouchDB system roles.
const (
	RoleAdmin      = "_admin"
	RoleReader     = "_reader"
	RoleWriter     = "_writer"
	RoleReplicator = "_replicator"
	RoleDBUpdates  = "_db_updates"
	RoleDesign     = "_design"
)

const typeJSON = "application/json"

// UserContext represents a [CouchDB UserContext object].
//
// [CouchDB UserContext object]: https://docs.couchdb.org/en/stable/json-structure.html#user-context-object
type UserContext struct {
	Database string   `json:"db,omitempty"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
	// Salt is needed to calculate cookie tokens.
	Salt string `json:"-"`
}

// HasRole returns true if the user has the specified role.
func (c *UserContext) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// Server is the interface for the server which exposes capabilities needed
// by auth handlers.
type Server interface {
	UserStore() UserStore
	Bind(*http.Request, interface{}) error
}

// AuthenticateFunc authenticates the HTTP request. On success, a user context
// must be returned. Any error will immediately terminate the authentication
// process, returning an error to the client. In particular, this means that
// an "unauthorized" error must not be returned if fallthrough is intended.
// If a response is sent, execution does not continue. This allows handlers
// to expose their own API endpoints (for example, the default cookie auth
// handler adds POST /_session and DELETE /_session handlers).
type AuthenticateFunc func(http.ResponseWriter, *http.Request) (*UserContext, error)

// Handler is an auth handler.
type Handler interface {
	// Init should return the name of the authenticatoin method, and an
	// authentication function. It is only called once on server startup.
	Init(Server) (string, AuthenticateFunc)
}
