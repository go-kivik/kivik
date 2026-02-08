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

package confadmin

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-kivik/kivik/v4"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
	"github.com/go-kivik/kivik/v4/x/kivikd/conf"
)

type confadmin struct {
	*conf.Conf
}

var _ authdb.UserStore = &confadmin{}

// New returns a new confadmin authentication service provider.
func New(c *conf.Conf) authdb.UserStore {
	return &confadmin{c}
}

func (c *confadmin) Validate(_ context.Context, username, password string) (*authdb.UserContext, error) {
	derivedKey, salt, iterations, err := c.getKeySaltIter(username)
	if err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			return nil, &internal.Error{Status: http.StatusUnauthorized, Message: "unauthorized"}
		}
		return nil, fmt.Errorf("unrecognized password hash: %w", err)
	}
	if !authdb.ValidatePBKDF2(password, salt, derivedKey, iterations) {
		return nil, &internal.Error{Status: http.StatusUnauthorized, Message: "unauthorized"}
	}
	return &authdb.UserContext{
		Name:  username,
		Roles: []string{"_admin"},
		Salt:  salt,
	}, nil
}

const hashPrefix = "-" + authdb.SchemePBKDF2 + "-"

func (c *confadmin) getKeySaltIter(username string) (key, salt string, iterations int, err error) {
	confName := "admins." + username
	if !c.IsSet(confName) {
		return "", "", 0, &internal.Error{Status: http.StatusNotFound, Message: "user not found"}
	}
	hash := c.GetString(confName)
	hash, ok := strings.CutPrefix(hash, hashPrefix)
	if !ok {
		return "", "", 0, errors.New("unrecognized password scheme")
	}
	const partsWanted = 3
	parts := strings.Split(hash, ",")
	if len(parts) != partsWanted {
		return "", "", 0, errors.New("unrecognized hash format")
	}
	if iterations, err = strconv.Atoi(parts[2]); err != nil {
		return "", "", 0, errors.New("unrecognized has format")
	}
	return parts[0], parts[1], iterations, nil
}

func (c *confadmin) UserCtx(_ context.Context, username string) (*authdb.UserContext, error) {
	_, salt, _, err := c.getKeySaltIter(username)
	if err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			return nil, &internal.Error{Status: http.StatusNotFound, Message: "user does not exist"}
		}
		return nil, fmt.Errorf("unrecognized password hash: %w", err)
	}
	return &authdb.UserContext{
		Name:  username,
		Roles: []string{"_admin"},
		Salt:  salt,
	}, nil
}
