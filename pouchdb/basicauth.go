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

import (
	"fmt"
	"strings"

	kivik "github.com/go-kivik/kivik/v4"
)

// BasicAuth handles HTTP Basic Auth for remote PouchDB connections. This
// is the only auth support built directly into PouchDB, so this is a very
// thin wrapper. This is the default authentication mechanism when credentials
// are provided in the connection URL, so this function is rarely ever needed.
func BasicAuth(username, password string) kivik.Option {
	return basicAuth{
		username: username,
		password: password,
	}
}

type basicAuth struct {
	username string
	password string
}

var _ kivik.Option = basicAuth{}

func (a basicAuth) Apply(target interface{}) {
	if client, ok := target.(*client); ok {
		client.setAuth(a.username, a.password)
	}
}

func (a basicAuth) String() string {
	return fmt.Sprintf("[BasicAuth{user:%s,pass:%s}]", a.username, strings.Repeat("*", len(a.password)))
}
