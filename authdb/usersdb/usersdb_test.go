package usersdb

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/authdb"
	_ "github.com/flimzy/kivik/driver/couchdb"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/test/kt"
)

type tuser struct {
	ID       string   `json:"_id"`
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Roles    []string `json:"roles"`
	Password string   `json:"password"`
}

var testUser = &tuser{
	ID:       "org.couchdb.user:testUsersdb",
	Name:     "testUsersdb",
	Type:     "user",
	Roles:    []string{"coolguy"},
	Password: "abc123",
}

func TestCouchAuth(t *testing.T) {
	client := kt.GetClient(t)
	db, err := client.DB(context.Background(), "_users")
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	rev, err := db.Put(context.Background(), testUser.ID, testUser)
	if err != nil {
		t.Fatalf("Failed to create user: %s", err)
	}
	defer db.Delete(context.Background(), testUser.ID, rev)
	auth := New(db)
	uCtx, err := auth.Validate(context.Background(), "testUsersdb", "abc123")
	if err != nil {
		t.Errorf("Validation failure for good password: %s", err)
	}
	if uCtx == nil {
		t.Errorf("User should have been validated")
	}
	uCtx, err = auth.Validate(context.Background(), "testUsersdb", "foobar")
	if errors.StatusCode(err) != kivik.StatusUnauthorized {
		t.Errorf("Expected Unauthorized password, got %s", err)
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

	uCtx, err = auth.UserCtx(context.Background(), "testUsersdb")
	if err != nil {
		t.Errorf("Failed to get roles for valid user: %s", err)
	}
	uCtx.Salt = "" // It's random, so remove it
	if !reflect.DeepEqual(uCtx, &authdb.UserContext{Name: "testUsersdb", Roles: []string{"coolguy"}}) {
		t.Errorf("Got unexpected output: %v", uCtx)
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
