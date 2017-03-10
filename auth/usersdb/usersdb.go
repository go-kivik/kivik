// Package usersdb provides auth facilities from a CouchDB _users database.
package usersdb

import (
	"context"
	"crypto/sha1"
	"fmt"

	"golang.org/x/crypto/pbkdf2"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/errors"
)

const userPrefix = "org.couchdb.user:"
const pbkdf2KeyLength = 20

type db struct {
	*kivik.DB
}

var _ auth.Handler = &db{}

// New returns a new auth.Handler backed by a the provided database.
func New(userDB *kivik.DB) auth.Handler {
	return &db{userDB}
}

type user struct {
	Roles          []string `json:"roles"`
	PasswordScheme string   `json:"password_scheme,omitempty"`
	Salt           string   `json:"salt,omitempty"`
	Iterations     int      `json:"iterations,omitempty"`
	DerivedKey     string   `json:"derived_key,omitempty"`
}

func (db *db) Validate(ctx context.Context, username, password string) (bool, error) {
	var u user
	if err := db.GetContext(ctx, userPrefix+username, &u, nil); err != nil {
		if errors.StatusCode(err) == errors.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	switch u.PasswordScheme {
	case "":
		return false, errors.New("No password scheme set for user")
	case "pbkdf2":
	default:
		return false, errors.Errorf("Unsupported password scheme: %s", u.PasswordScheme)
	}
	key := fmt.Sprintf("%x", pbkdf2.Key([]byte(password), []byte(u.Salt), u.Iterations, pbkdf2KeyLength, sha1.New))
	return key == u.DerivedKey, nil
}

func (db *db) Roles(ctx context.Context, username string) ([]string, error) {
	var u user
	return u.Roles, db.GetContext(ctx, userPrefix+username, &u, nil)
}
