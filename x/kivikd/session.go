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

package kivikd

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4/x/kivikd/auth"
)

// DefaultInsecureSecret is the hash secret used if couch_httpd_auth.secret
// is unconfigured. Please configure couch_httpd_auth.secret, or they're all
// gonna laugh at you!
const DefaultInsecureSecret = "They're all gonna laugh at you!"

// DefaultSessionTimeout is the default session timeout, in seconds, used if
// couch_httpd_auth.timeout is inuset.
const DefaultSessionTimeout = 600

func (s *Service) getAuthSecret() string {
	if s.Conf().IsSet("couch_httpd_auth.secret") {
		return s.Conf().GetString("couch_httpd_auth.secret")
	}
	return DefaultInsecureSecret
}

func setSession() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()
			// We store a pointer to a pointer, so the underlying pointer can
			// be updated by the auth process, without losing the reference.
			session := &auth.Session{}
			ctx = context.WithValue(ctx, SessionKey, &session)
			r = r.WithContext(ctx)
			next.ServeHTTP(w, r)
		})
	}
}
