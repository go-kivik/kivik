// Package authgroup groups two or more authentication backends together, trying
// one, then falling through to the others.
package authgroup

import (
	"context"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/errors"
)

// AuthGroup is a group of auth handlers, to be tried in turn.
type AuthGroup []auth.Handler

var _ auth.Handler = AuthGroup{}

// New initializes a group of auth handlers. Each one is tried in turn, in the
// order passed to New.
func New(authHandlers ...auth.Handler) auth.Handler {
	return append(AuthGroup{}, authHandlers...)
}

// Validate loops through each of the auth handlers, in the order passed to New,
// until the context is cancelled, in which case the context's error is returned.
// The first validation success returns. Errors are discarded unless all auth
// handlers fail to validate the user, in which case only the first error
// received will be returned.
func (g AuthGroup) Validate(ctx context.Context, username, password string) (bool, error) {
	var firstErr error
	for _, handler := range g {
		valid, err := handler.Validate(ctx, username, password)
		if valid && err == nil {
			return true, nil
		}
		if err != nil {
			if firstErr == nil {
				firstErr = err
			}
		}
		select {
		// See if our context has expired
		case <-ctx.Done():
			return false, ctx.Err()
		default:
		}
	}
	return false, firstErr
}

// Roles loops through each of the auth handlers, in the order passed to New
// until the context is cancelled, in which case the context's error is returned.
// The first one to not return an error returns. If all of the handlers return
// a Not Found error, Not Found is returned. If any other errors are returned,
// the first is returned to the caller.
func (g AuthGroup) Roles(ctx context.Context, username string) ([]string, error) {
	var firstErr error
	for _, handler := range g {
		roles, err := handler.Roles(ctx, username)
		if err == nil {
			return roles, nil
		}
		if errors.StatusCode(err) != kivik.StatusNotFound && firstErr == nil {
			firstErr = err
		}
		select {
		// See if our context has expired
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}
	}
	if firstErr == nil {
		return nil, errors.Status(kivik.StatusNotFound, "user not found")
	}
	return nil, firstErr
}
