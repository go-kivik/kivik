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
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/server/auth"
)

type userStores []auth.UserStore

var _ auth.UserStore = userStores{}

func (s userStores) Validate(ctx context.Context, username, password string) (*auth.UserContext, error) {
	for _, store := range s {
		userCtx, err := store.Validate(ctx, username, password)
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			continue
		}
		return userCtx, err
	}
	return nil, &internal.Error{Status: http.StatusUnauthorized, Message: "Invalid username or password"}
}

func (s userStores) UserCtx(ctx context.Context, username string) (*auth.UserContext, error) {
	for _, store := range s {
		userCtx, err := store.UserCtx(ctx, username)
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			continue
		}
		return userCtx, err
	}
	return nil, &internal.Error{Status: http.StatusNotFound, Message: "User not found"}
}
