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
	"context"
	"net/http"
	"sync"

	"github.com/go-kivik/kivik/v4/internal"
)

// A UserStore provides an AuthHandler with access to a user store for.
type UserStore interface {
	// Validate returns a user context object if the credentials are valid. An
	// error must be returned otherwise. A Not-Found error will continue to the
	// next user store, while any other error will terminate the auth process.
	Validate(ctx context.Context, username, password string) (user *UserContext, err error)
	// UserCtx returns a user context object if the user exists. It is used by
	// AuthHandlers that don't validate the password (e.g. Cookie auth). If the
	// user does not exist, a Not-Found error will be returned.
	UserCtx(ctx context.Context, username string) (user *UserContext, err error)
}

// MemoryUserStore is a simple in-memory user store.
type MemoryUserStore struct {
	users sync.Map
}

var _ UserStore = (*MemoryUserStore)(nil)

type memoryUser struct {
	Password string
	Roles    []string
}

// NewMemoryUserStore returns a new MemoryUserStore.
func NewMemoryUserStore() *MemoryUserStore {
	return &MemoryUserStore{}
}

// AddUser adds a user to the store. It returns an error if the user already
// exists.
func (s *MemoryUserStore) AddUser(username, password string, roles []string) error {
	_, loaded := s.users.LoadOrStore(username, &memoryUser{
		Password: password,
		Roles:    roles,
	})
	if loaded {
		return &internal.Error{Status: http.StatusConflict, Message: "User already exists"}
	}
	return nil
}

// DeleteUser deletes a user from the store.
func (s *MemoryUserStore) DeleteUser(username string) {
	s.users.Delete(username)
}

var (
	errNotFound     = &internal.Error{Status: http.StatusNotFound, Message: "User not found"}
	errUnauthorized = &internal.Error{Status: http.StatusUnauthorized, Message: "Invalid username or password"}
)

// Validate returns a user context object if the credentials are valid.
func (s *MemoryUserStore) Validate(_ context.Context, username, password string) (*UserContext, error) {
	user, ok := s.users.Load(username)
	if !ok {
		return nil, errNotFound
	}
	if user.(*memoryUser).Password != password {
		return nil, errUnauthorized
	}
	return &UserContext{
		Name:  username,
		Roles: user.(*memoryUser).Roles,
	}, nil
}

// UserCtx returns a user context object if the user exists.
func (s *MemoryUserStore) UserCtx(_ context.Context, username string) (*UserContext, error) {
	user, ok := s.users.Load(username)
	if !ok {
		return nil, errNotFound
	}
	return &UserContext{
		Name:  username,
		Roles: user.(*memoryUser).Roles,
	}, nil
}
