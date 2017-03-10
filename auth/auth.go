// Package auth provides a standard interface for user credential validation
// by a Kivik server.
package auth

import "context"

// A Handler is used by a server to validate auth credentials.
type Handler interface {
	// Validate returns true if the credentials are valid, false otherwise.
	// Validate must not return a Not Found error if the user does not exist,
	// rather returning a false validation.
	Validate(ctx context.Context, username, password string) (ok bool, err error)
	// Roles returns the roles to which the user belongs. Roles must return
	// a Not Found error if the user does not exist.
	Roles(ctx context.Context, username string) (roles []string, err error)
}

// PBKDF2KeyLength is the key length, in bytes, of the PBKDF2 keys used by
// CouchDB.
const PBKDF2KeyLength = 20

// SchemePBKDF2 is the default CouchDB password scheme.
const SchemePBKDF2 = "pbkdf2"
