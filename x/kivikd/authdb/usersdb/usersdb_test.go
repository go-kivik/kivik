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

package usersdb

import (
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
)

type tuser struct {
	ID       string   `json:"_id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Roles    []string `json:"roles"`
	Password string   `json:"password"`
}

func TestCouchAuth(t *testing.T) {
	t.Skip("Reconfigure test not to require Docker")
	client := kt.GetClient(t)
	db := client.DB("_users")
	if err := db.Err(); err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	name := kt.TestDBName(t)
	user := &tuser{
		ID:       kivik.UserPrefix + name,
		Name:     name,
		Type:     "user",
		Roles:    []string{"coolguy"},
		Password: "abc123",
	}
	rev, err := db.Put(t.Context(), user.ID, user)
	if err != nil {
		t.Fatalf("Failed to create user: %s", err)
	}
	defer db.Delete(t.Context(), user.ID, rev) // nolint:errcheck
	auth := New(db)
	t.Run("sync", func(t *testing.T) {
		t.Run("Validate", func(t *testing.T) {
			t.Parallel()
			t.Run("ValidUser", func(t *testing.T) {
				uCtx, err := auth.Validate(t.Context(), user.Name, "abc123")
				if err != nil {
					t.Errorf("Validation failure for good password: %s", err)
				}
				if uCtx == nil {
					t.Errorf("User should have been validated")
				}
			})
			t.Run("WrongPassword", func(t *testing.T) {
				uCtx, err := auth.Validate(t.Context(), user.Name, "foobar")
				if kivik.HTTPStatus(err) != http.StatusUnauthorized {
					t.Errorf("Expected Unauthorized password, got %s", err)
				}
				if uCtx != nil {
					t.Errorf("User should not have been validated with wrong password")
				}
			})
			t.Run("MissingUser", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.Validate(t.Context(), "nobody", "foo")
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
			t.Run("ValidUser", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.UserCtx(t.Context(), user.Name)
				if err != nil {
					t.Errorf("Failed to get roles: %s", err)
				}
				uCtx.Salt = "" // It's random, so remove it
				if !reflect.DeepEqual(uCtx, &authdb.UserContext{Name: user.Name, Roles: []string{"coolguy"}}) {
					t.Errorf("Got unexpected output: %v", uCtx)
				}
			})
			t.Run("MissingUser", func(t *testing.T) {
				t.Parallel()
				_, err := auth.UserCtx(t.Context(), "nobody")
				if kivik.HTTPStatus(err) != http.StatusNotFound {
					var msg string
					if err != nil {
						msg = fmt.Sprintf(" Got: %s", err)
					}
					t.Errorf("Expected Not Found fetching roles: %s", msg)
				}
			})
		})
	})
}
