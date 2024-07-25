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

package server

import (
	"github.com/go-kivik/kivik/v4/x/server/auth"
	"github.com/go-kivik/kivik/v4/x/server/config"
)

// Option is a server option.
type Option interface {
	apply(*Server)
}

type authHandlerOption []auth.Handler

func (h authHandlerOption) apply(s *Server) {
	for _, handler := range h {
		_, auth := handler.Init(&authService{s})
		s.authFuncs = append(s.authFuncs, auth)
	}
}

// WithAuthHandlers adds the provided auth handlers to the server. May be
// specified more than once. Order is significant. Each auth request is passed
// through each handler in the order specified, until one returns a user
// context or an error. If no handlers are specified, the server will operate
// as a PERPETUAL ADMIN PARTY!
func WithAuthHandlers(h ...auth.Handler) Option {
	return authHandlerOption(h)
}

type userStoreOption []auth.UserStore

func (s userStoreOption) apply(srv *Server) {
	for _, store := range s {
		srv.userStores = append(srv.userStores, store)
	}
}

// WithUserStores adds the provided user stores to the server. May be specified
// more than once. Order is significant. Each user store is queried in the order
// specified, until one returns a user context or an error.
func WithUserStores(us ...auth.UserStore) Option {
	return userStoreOption(us)
}

type configOption [1]config.Config

func (c configOption) apply(s *Server) {
	s.config = c[0]
}

// WithConfig sets the server configuration. If not set,
// [github.com/go-kivik/kivik/v4/x/server/config.Default()] will be used.
func WithConfig(c config.Config) Option {
	return configOption{c}
}
