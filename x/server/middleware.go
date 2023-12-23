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
// +build !js

package server

import (
	"io"
	"net/http"

	"github.com/go-chi/chi/v5"
)

// GetHead automatically route undefined HEAD requests to GET handlers, and
// discards the response body for such requests.
//
// Forked from chi middleware package.
func GetHead(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "HEAD" {
			rctx := chi.RouteContext(r.Context())
			routePath := rctx.RoutePath
			if routePath == "" {
				if r.URL.RawPath != "" {
					routePath = r.URL.RawPath
				} else {
					routePath = r.URL.Path
				}
			}

			// Temporary routing context to look-ahead before routing the request
			tctx := chi.NewRouteContext()

			// Attempt to find a HEAD handler for the routing path, if not found, traverse
			// the router as through its a GET route, but proceed with the request
			// with the HEAD method.
			if !rctx.Routes.Match(tctx, "HEAD", routePath) {
				type httpWriter interface { // Just the methods unique to http.ResponseWriter
					Header() http.Header
					WriteHeader(statusCode int)
				}
				discardWriter := struct {
					httpWriter
					io.Writer
				}{
					httpWriter: w,
					Writer:     io.Discard,
				}
				rctx.RouteMethod = "GET"
				rctx.RoutePath = routePath
				next.ServeHTTP(discardWriter, r)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}
