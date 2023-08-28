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

package chttp

import (
	"net/http/cookiejar"

	"golang.org/x/net/publicsuffix"
)

// Authenticator is an interface that provides authentication to a server.
type Authenticator interface {
	Authenticate(*Client) error
}

func (a *CookieAuth) setCookieJar() {
	// If a jar is already set, just use it
	if a.client.Jar != nil {
		return
	}
	// cookiejar.New never returns an error
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	a.client.Jar = jar
}
