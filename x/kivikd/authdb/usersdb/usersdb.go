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

// Package usersdb provides auth facilities from a CouchDB _users database.
package usersdb

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"

	"golang.org/x/crypto/pbkdf2"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
)

type db struct {
	*kivik.DB
}

var _ authdb.UserStore = &db{}

// New returns a new authdb.UserStore backed by a the provided database.
func New(userDB *kivik.DB) authdb.UserStore {
	return &db{userDB}
}

type user struct {
	Name           string   `json:"name"`
	Roles          []string `json:"roles"`
	PasswordScheme string   `json:"password_scheme,omitempty"`
	Salt           string   `json:"salt,omitempty"`
	Iterations     int      `json:"iterations,omitempty"`
	DerivedKey     string   `json:"derived_key,omitempty"`
}

func (db *db) getUser(ctx context.Context, username string) (*user, error) {
	var u user
	if err := db.Get(ctx, kivik.UserPrefix+username, nil).ScanDoc(&u); err != nil {
		return nil, err
	}
	return &u, nil
}

func (db *db) Validate(ctx context.Context, username, password string) (*authdb.UserContext, error) {
	u, err := db.getUser(ctx, username)
	if err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			err = &internal.Error{Status: http.StatusUnauthorized, Message: "unauthorized"}
		}
		return nil, err
	}

	switch u.PasswordScheme {
	case "":
		return nil, errors.New("no password scheme set for user")
	case authdb.SchemePBKDF2:
	default:
		return nil, fmt.Errorf("unsupported password scheme: %s", u.PasswordScheme)
	}
	key := hex.EncodeToString(pbkdf2.Key([]byte(password), []byte(u.Salt), u.Iterations, authdb.PBKDF2KeyLength, sha1.New))
	if key != u.DerivedKey {
		return nil, &internal.Error{Status: http.StatusUnauthorized, Message: "unauthorized"}
	}
	return &authdb.UserContext{
		Name:  u.Name,
		Roles: u.Roles,
		Salt:  u.Salt,
	}, nil
}

func (db *db) UserCtx(ctx context.Context, username string) (*authdb.UserContext, error) {
	u, err := db.getUser(ctx, username)
	if err != nil {
		return nil, err
	}
	return &authdb.UserContext{
		Name:  u.Name,
		Roles: u.Roles,
		Salt:  u.Salt,
	}, nil
}
