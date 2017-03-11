// Package authdb provides a standard interface to an authentication user store
// to be used by AuthHandlers.
package authdb

import (
	"context"
	"crypto/sha1"
	"fmt"

	"golang.org/x/crypto/pbkdf2"
)

// A UserStore provides an AuthHandler with access to a user store for.
type UserStore interface {
	// Validate returns a user context object if the credentials are valid. An
	// error must be returned otherwise. A Not Found error must not be returned.
	// Not Found should be treated identically to Unauthorized.
	Validate(ctx context.Context, username, password string) (user *UserContext, err error)
	// UserCtx returns a user context object if the user exists. It is used by
	// AuthHandlers that don't validate the password (e.g. Cookie auth).
	UserCtx(ctx context.Context, username string) (user *UserContext, err error)
}

// PBKDF2KeyLength is the key length, in bytes, of the PBKDF2 keys used by
// CouchDB.
const PBKDF2KeyLength = 20

// SchemePBKDF2 is the default CouchDB password scheme.
const SchemePBKDF2 = "pbkdf2"

// UserContext represents a CouchDB UserContext object.
// See http://docs.couchdb.org/en/2.0.0/json-structure.html#userctx-object.
type UserContext struct {
	Database     string   `json:"db,omitempty"`
	AuthDatabase string   `json:"authentication_db,omitempty"`
	Name         string   `json:"name"`
	Roles        []string `json:"roles"`
}

// ValidatePBKDF2 returns true if the calculated hash matches the derivedKey.
func ValidatePBKDF2(password, salt, derivedKey string, iterations int) bool {
	hash := fmt.Sprintf("%x", pbkdf2.Key([]byte(password), []byte(salt), iterations, PBKDF2KeyLength, sha1.New))
	return hash == derivedKey
}
