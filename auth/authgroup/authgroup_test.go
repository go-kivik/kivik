package authgroup

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth/confadmin"
	"github.com/flimzy/kivik/auth/usersdb"
	"github.com/flimzy/kivik/config"
	_ "github.com/flimzy/kivik/driver/couchdb"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/serve/config/memconf"
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

func TestConfAdminAuth(t *testing.T) {
	// Set up first auth backend
	conf1 := config.New(memconf.New())
	conf1.Set("admins", "bob", "-pbkdf2-792221164f257de22ad72a8e94760388233e5714,7897f3451f59da741c87ec5f10fe7abe,10")
	auth1 := confadmin.New(conf1)

	// Set up second auth backend
	client := kt.GetClient(t)
	db, err := client.DB("_users")
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	rev, err := db.Put(testUser.ID, testUser)
	if err != nil {
		t.Fatalf("Failed to create user: %s", err)
	}
	defer db.Delete(testUser.ID, rev)
	auth2 := usersdb.New(db)

	auth := New(auth1, auth2)

	valid, err := auth.Validate(kt.CTX, "bob", "abc123")
	if err != nil {
		t.Errorf("Validation failure for bob/good password: %s", err)
	}
	if !valid {
		t.Errorf("User should have been validated")
	}

	notValid, err := auth.Validate(kt.CTX, "bob", "foobar")
	if err != nil {
		t.Errorf("Validation failure for bob/bad password: %s", err)
	}
	if notValid {
		t.Errorf("User should not have been validated with wrong password")
	}

	valid, err = auth.Validate(kt.CTX, "test", "abc123")
	if err != nil {
		t.Errorf("Validation failure for good password: %s", err)
	}
	if !valid {
		t.Errorf("User should have been validated")
	}
	notValid, err = auth.Validate(kt.CTX, "test", "foobar")
	if err != nil {
		t.Errorf("Validation failure for bad password: %s", err)
	}
	if notValid {
		t.Errorf("User should not have been validated with wrong password")
	}
	notValid, err = auth.Validate(kt.CTX, "nobody", "foo")
	if err != nil {
		t.Errorf("Validation failure for bad username: %s", err)
	}
	if notValid {
		t.Errorf("User should not have been validated with wrong username")
	}

	roles, err := auth.Roles(kt.CTX, "test")
	if err != nil {
		t.Errorf("Failed to get roles for valid user: %s", err)
	}
	if !reflect.DeepEqual(roles, []string{"coolguy"}) {
		t.Errorf("Got unexpected roles: %v", roles)
	}

	roles, err = auth.Roles(kt.CTX, "bob")
	if err != nil {
		t.Errorf("Failed to get roles for valid user: %s", err)
	}
	if !reflect.DeepEqual(roles, []string{"_admin"}) {
		t.Errorf("Got unexpected roles: %v", roles)
	}

	_, err = auth.Roles(kt.CTX, "nobody")
	if errors.StatusCode(err) != kivik.StatusNotFound {
		var msg string
		if err != nil {
			msg = fmt.Sprintf(" Got: %s", err)
		}
		t.Errorf("Expected Not Found fetching roles for bad username.%s", msg)
	}
}
