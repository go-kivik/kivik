package couchauth

import (
	"testing"

	_ "github.com/flimzy/kivik/driver/couchdb"
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
	if _, err := New("http://foo.com:port with spaces/"); err == nil {
		t.Errorf("Expected error for invalid URL.")
	}
	if _, err := New("http://foo:bar@foo.com/"); err == nil {
		t.Error("Expected error for DSN with credentials.")
	}
}

func TestCouchAuth(t *testing.T) {
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
	auth, err := New(kt.NoAuthDSN(t))
	if err != nil {
		t.Fatalf("Failed to connect to remote server: %s", err)
	}
	valid, err := auth.Validate(kt.CTX, "test", "abc123")
	if err != nil {
		t.Errorf("Validation failure for good password: %s", err)
	}
	if !valid {
		t.Errorf("User should have been validated")
	}
	notValid, err := auth.Validate(kt.CTX, "test", "foobar")
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

	// roles, err := auth.Roles(kt.CTX, "test")
	// if err != nil {
	// 	t.Errorf("Failed to get roles for valid user: %s", err)
	// }
	// if !reflect.DeepEqual(roles, []string{"coolguy"}) {
	// 	t.Errorf("Got unexpected roles.")
	// }
	// _, err = auth.Roles(kt.CTX, "nobody")
	// if errors.StatusCode(err) != kivik.StatusNotFound {
	// 	var msg string
	// 	if err != nil {
	// 		msg = fmt.Sprintf(" Got: %s", err)
	// 	}
	// 	t.Errorf("Expected Not Found fetching roles for bad username.%s", msg)
	// }
}
