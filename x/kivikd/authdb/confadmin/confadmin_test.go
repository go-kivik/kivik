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

package confadmin

import (
	"context"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/spf13/viper"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb"
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
	"github.com/go-kivik/kivik/v4/x/kivikd/conf"
)

func TestInvalidHashes(t *testing.T) {
	c := &conf.Conf{Viper: viper.New()}
	c.Set("admins.test", "-pbkXXdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,10")
	auth := New(c)
	if _, err := auth.Validate(context.Background(), "test", "123"); err == nil {
		t.Errorf("Expected error for invalid scheme")
	}
	if _, err := auth.Validate(context.Background(), "test", "123"); err == nil {
		t.Errorf("Expected error for too many commas")
	}
	c.Set("admins.test", "-pbkdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,pig")
	if _, err := auth.Validate(context.Background(), "test", "123"); err == nil {
		t.Errorf("Expected error for invalid iterations integer")
	}
}

func TestConfAdminAuth(t *testing.T) {
	c := &conf.Conf{Viper: viper.New()}
	c.Set("admins.test", "-pbkdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,10")
	auth := New(c)

	t.Run("sync", func(t *testing.T) {
		t.Run("Validate", func(t *testing.T) {
			t.Parallel()
			t.Run("ValidUser", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.Validate(context.Background(), "test", "abc123")
				if err != nil {
					t.Errorf("Validation failure for good password: %s", err)
				}
				if uCtx == nil {
					t.Errorf("User should have been validated")
				}
			})
			t.Run("WrongPassword", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.Validate(context.Background(), "test", "foobar")
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
			t.Run("ValidUser", func(t *testing.T) {
				t.Parallel()
				uCtx, err := auth.UserCtx(context.Background(), "test")
				if err != nil {
					t.Errorf("Failed to get roles for valid user: %s", err)
				}
				if !reflect.DeepEqual(uCtx, &authdb.UserContext{Name: "test", Roles: []string{"_admin"}, Salt: "7897f3451f59da741c87ec5f10fe7abe"}) {
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
