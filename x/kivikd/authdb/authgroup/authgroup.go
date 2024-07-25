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

package authgroup

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
)

// AuthGroup is a group of auth handlers, to be tried in turn.
type AuthGroup []authdb.UserStore

var _ authdb.UserStore = AuthGroup{}

// New initializes a group of auth handlers. Each one is tried in turn, in the
// order passed to New.
func New(userStores ...authdb.UserStore) authdb.UserStore {
	return append(AuthGroup{}, userStores...)
}

func (g AuthGroup) loop(ctx context.Context, fn func(authdb.UserStore) (*authdb.UserContext, error)) (*authdb.UserContext, error) {
	var firstErr error
	for _, store := range g {
		uCtx, err := fn(store)
		if err == nil {
			return uCtx, nil
		}
		if kivik.HTTPStatus(err) != http.StatusNotFound && firstErr == nil {
			firstErr = err
		}
		select {
		// See if our context has expired
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	if firstErr == nil {
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "user not found"}
	}
	return nil, firstErr
}

// Validate loops through each of the auth handlers, in the order passed to New,
// until the context is cancelled, in which case the context's error is returned.
// The first validation success returns. Errors are discarded unless all auth
// handlers fail to validate the user, in which case only the first error
// received will be returned.
func (g AuthGroup) Validate(ctx context.Context, username, password string) (*authdb.UserContext, error) {
	return g.loop(ctx, func(store authdb.UserStore) (*authdb.UserContext, error) {
		return store.Validate(ctx, username, password)
	})
}

// UserCtx loops through each of the auth handlers, in the order passed to New
// until the context is cancelled, in which case the context's error is returned.
// The first one to not return an error returns. If all of the handlers return
// a Not Found error, Not Found is returned. If any other errors are returned,
// the first is returned to the caller.
func (g AuthGroup) UserCtx(ctx context.Context, username string) (*authdb.UserContext, error) {
	return g.loop(ctx, func(store authdb.UserStore) (*authdb.UserContext, error) {
		return store.UserCtx(ctx, username)
	})
}
