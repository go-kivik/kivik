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
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb/confadmin"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb/usersdb"
	"github.com/go-kivik/kivik/v4/x/kivikd/conf"
)

type tuser struct {
	ID       string   `json:"_id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Roles    []string `json:"roles"`
	Password string   `json:"password"`
}

func TestConfAdminAuth(t *testing.T) {
	t.Skip("Reconfigure test not to require Docker")
	// Set up first auth backend
	c1 := conf.New()
	c1.Set("admins.bob", "-pbkdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,10")
	auth1 := confadmin.New(c1)

	// Set up second auth backend
	client := kt.GetClient(t)
	db := client.DB("_users")
	if e := db.Err(); e != nil {
		t.Fatalf("Failed to connect to db: %s", e)
	}
	name := kt.TestDBName(t)
	user := &tuser{
		ID:       kivik.UserPrefix + name,
		Name:     name,
		Type:     "user",
		Roles:    []string{"coolguy"},
		Password: "abc123",
	}
	rev, e := db.Put(context.Background(), user.ID, user)
	if e != nil {
		t.Fatalf("Failed to create user: %s", e)
	}
	defer db.Delete(context.Background(), user.ID, rev) // nolint: errcheck
	auth2 := usersdb.New(db)

	auth := New(auth1, auth2)

	t.Run("sync", func(t *testing.T) {
		t.Run("Validate", func(t *testing.T) {
			t.Parallel()
			t.Run("BobValid", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.Validate(context.Background(), "bob", "abc123")
				if err != nil {
					t.Errorf("Validation failure for bob/good password: %s", err)
				}
				if uCtx == nil {
					t.Errorf("User should have been validated")
				}
			})
			t.Run("BobInvalid", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.Validate(context.Background(), "bob", "foobar")
				if kivik.HTTPStatus(err) != http.StatusUnauthorized {
					t.Errorf("Expected Unauthorized for bad password, got %s", err)
				}
				if uCtx != nil {
					t.Errorf("User should not have been validated with wrong password")
				}
			})
			t.Run("TestUserValid", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.Validate(context.Background(), user.Name, "abc123")
				if err != nil {
					t.Errorf("Validation failure for good password: %s", err)
				}
				if uCtx == nil {
					t.Errorf("User should have been validated")
				}
			})
			t.Run("TestUserInvalid", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.Validate(context.Background(), user.Name, "foobar")
				if kivik.HTTPStatus(err) != http.StatusUnauthorized {
					t.Errorf("Expected Unauthorized for bad password, got %s", err)
				}
				if uCtx != nil {
					t.Errorf("User should not have been validated with wrong password")
				}
			})
			t.Run("MissingUser", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.Validate(context.Background(), "nobody", "foo")
				if kivik.HTTPStatus(err) != http.StatusUnauthorized {
					t.Errorf("Expected Unauthorized for bad username, got %s", err)
				}
				if uCtx != nil {
					t.Errorf("User should not have been validated with wrong username")
				}
			})
		})
		t.Run("Context", func(t *testing.T) {
			t.Parallel()
			t.Run("TestUser", func(t *testing.T) {
				uCtx, err := auth.UserCtx(context.Background(), user.Name)
				if err != nil {
					t.Errorf("Failed to get roles for valid user: %s", err)
				}
				uCtx.Salt = "" // It's random, so don't fail if it doesn't match
				if !reflect.DeepEqual(uCtx, &authdb.UserContext{Name: user.Name, Roles: []string{"coolguy"}}) {
					t.Errorf("Got unexpected context: %v", uCtx)
				}
			})
			t.Run("Bob", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.UserCtx(context.Background(), "bob")
				if err != nil {
					t.Errorf("Failed to get roles for valid user: %s", err)
				}
				if !reflect.DeepEqual(uCtx, &authdb.UserContext{Name: "bob", Roles: []string{"_admin"}, Salt: "7897f3451f59da741c87ec5f10fe7abe"}) {
					t.Errorf("Got unexpected context: %v", uCtx)
				}
			})
			t.Run("MissingUser", func(t *testing.T) {
				_, err := auth.UserCtx(context.Background(), "nobody")
				if kivik.HTTPStatus(err) != http.StatusNotFound {
					var msg string
					if err != nil {
						msg = fmt.Sprintf(" Got: %s", err)
					}
					t.Errorf("Expected Not Found fetching roles for bad username.%s", msg)
				}
			})
		})
	})
}
