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

package pouchdb

// BasicAuth handles HTTP Basic Auth for remote PouchDB connections. This
// is the only auth support built directly into PouchDB, so this is a very
// thin wrapper.
type BasicAuth struct {
	Name     string
	Password string
}

// authenticate sets the HTTP Basic Auth parameters.
func (a *BasicAuth) authenticate(c *client) error {
	c.opts["authenticator"] = Options{
		"auth": map[string]interface{}{
			"username": a.Name,
			"password": a.Password,
		},
	}
	return nil
}
