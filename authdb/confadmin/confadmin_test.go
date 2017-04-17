package confadmin

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/authdb"
	"github.com/flimzy/kivik/config"
	_ "github.com/flimzy/kivik/driver/couchdb"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/serve/config/memconf"
)

type tuser struct {
	ID       string   `json:"_id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Roles    []string `json:"roles"`
	Password string   `json:"password"`
}

var testUser = &tuser{
	ID:       "org.couchdb.user:test",
	Name:     "test",
	Type:     "user",
	Roles:    []string{"coolguy"},
	Password: "abc123",
}

func TestInvalidHashes(t *testing.T) {
	conf := config.New(memconf.New())
	auth := New(conf)
	conf.Set(context.Background(), "admins", "test", "-pbkXXdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,10")
	if _, err := auth.Validate(context.Background(), "test", "123"); err == nil {
		t.Errorf("Expected error for invalid scheme")
	}
	conf.Set(context.Background(), "admins", "test", "-pbkdf2-792221164f257de22ad72a8e,94760388233e5714,7897f345,1f59da741c87ec5f10fe7abe,10")
	if _, err := auth.Validate(context.Background(), "test", "123"); err == nil {
		t.Errorf("Expected error for too many commas")
	}
	conf.Set(context.Background(), "admins", "test", "-pbkdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,pig")
	if _, err := auth.Validate(context.Background(), "test", "123"); err == nil {
		t.Errorf("Expected error for invalid iterations integer")
	}
}

func TestConfAdminAuth(t *testing.T) {
	conf := config.New(memconf.New())
	auth := New(conf)

	conf.Set(context.Background(), "admins", "test", "-pbkdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,10")
	uCtx, err := auth.Validate(context.Background(), "test", "abc123")
	if err != nil {
		t.Errorf("Validation failure for good password: %s", err)
	}
	if uCtx == nil {
		t.Errorf("User should have been validated")
	}
	uCtx, err = auth.Validate(context.Background(), "test", "foobar")
	if errors.StatusCode(err) != kivik.StatusUnauthorized {
		t.Errorf("Expected Unauthorized for bad password, got %s", err)
	}
	if uCtx != nil {
		t.Errorf("User should not have been validated with wrong password")
	}
	uCtx, err = auth.Validate(context.Background(), "nobody", "foo")
	if errors.StatusCode(err) != kivik.StatusUnauthorized {
		t.Errorf("Expected Unauthorized for bad username, got %s", err)
	}
	if uCtx != nil {
		t.Errorf("User should not have been validated with wrong username")
	}

	uCtx, err = auth.UserCtx(context.Background(), "test")
	if err != nil {
		t.Errorf("Failed to get roles for valid user: %s", err)
	}
	if !reflect.DeepEqual(uCtx, &authdb.UserContext{Name: "test", Roles: []string{"_admin"}, Salt: "7897f3451f59da741c87ec5f10fe7abe"}) {
		t.Errorf("Got unexpected context: %v", uCtx)
	}
	_, err = auth.UserCtx(context.Background(), "nobody")
	if errors.StatusCode(err) != kivik.StatusNotFound {
		var msg string
		if err != nil {
			msg = fmt.Sprintf(" Got: %s", err)
		}
		t.Errorf("Expected Not Found fetching roles for bad username.%s", msg)
	}
}
