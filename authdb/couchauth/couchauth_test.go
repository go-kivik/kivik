package couchauth

import (
	"context"
	"testing"

	"github.com/flimzy/kivik"
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
	ID:       "org.couchdb.user:test",
	Name:     "test",
	Type:     "user",
	Roles:    []string{"coolguy"},
	Password: "abc123",
}

func TestBadDSN(t *testing.T) {
	if _, err := New(context.Background(), "http://foo.com:port with spaces/"); err == nil {
		t.Errorf("Expected error for invalid URL.")
	}
	if _, err := New(context.Background(), "http://foo:bar@foo.com/"); err == nil {
		t.Error("Expected error for DSN with credentials.")
	}
}

func TestCouchAuth(t *testing.T) {
	client := kt.GetClient(t)
	db, err := client.DB(context.Background(), "_users")
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	// Courtesy flush
	kt.DeleteUser(db, testUser.ID, t)
	rev, err := db.Put(context.Background(), testUser.ID, testUser)
	if err != nil {
		t.Fatalf("Failed to create user: %s", err)
	}
	defer db.Delete(context.Background(), testUser.ID, rev)
	auth, err := New(context.Background(), kt.NoAuthDSN(t))
	if err != nil {
		t.Fatalf("Failed to connect to remote server: %s", err)
	}
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

	// roles, err := auth.Roles(context.Background(), "test")
	// if err != nil {
	// 	t.Errorf("Failed to get roles for valid user: %s", err)
	// }
	// if !reflect.DeepEqual(roles, []string{"coolguy"}) {
	// 	t.Errorf("Got unexpected roles.")
	// }
	// _, err = auth.Roles(context.Background(), "nobody")
	// if errors.StatusCode(err) != kivik.StatusNotFound {
	// 	var msg string
	// 	if err != nil {
	// 		msg = fmt.Sprintf(" Got: %s", err)
	// 	}
	// 	t.Errorf("Expected Not Found fetching roles for bad username.%s", msg)
	// }
}
