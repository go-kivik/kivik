// Package confadmin provides an authentication service for admins configured
// in server configuration.
package confadmin

import (
	"context"
	"strconv"
	"strings"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/authdb"
	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/errors"
)

type conf struct {
	*config.Config
}

var _ authdb.UserStore = &conf{}

// New returns a new confadmin authentication service provider.
func New(c *config.Config) authdb.UserStore {
	return &conf{c}
}

func (c *conf) Validate(ctx context.Context, username, password string) (*authdb.UserContext, error) {
	hash, err := c.GetContext(ctx, "admins", username)
	if err != nil {
		if errors.StatusCode(err) == kivik.StatusNotFound {
			err = errors.Status(kivik.StatusUnauthorized, "unauthorized")
		}
		return nil, err
	}
	derivedKey, salt, iterations, err := keySaltIter(hash)
	if err != nil {
		return nil, errors.Wrap(err, "unrecognized password hash")
	}
	if !authdb.ValidatePBKDF2(password, salt, derivedKey, iterations) {
		return nil, errors.Status(kivik.StatusUnauthorized, "unauthorized")
	}
	return &authdb.UserContext{
		Name:  username,
		Roles: []string{"_admin"},
	}, nil
}

const hashPrefix = "-" + authdb.SchemePBKDF2 + "-"

func keySaltIter(hash string) (key, salt string, iterations int, err error) {
	if !strings.HasPrefix(hash, hashPrefix) {
		return "", "", 0, errors.New("unrecognized password scheme")
	}
	parts := strings.Split(strings.TrimPrefix(hash, hashPrefix), ",")
	if len(parts) != 3 {
		return "", "", 0, errors.New("unrecognized hash format")
	}
	if iterations, err = strconv.Atoi(parts[2]); err != nil {
		return "", "", 0, errors.New("unrecognized has format")
	}
	return parts[0], parts[1], iterations, nil
}

func (c *conf) UserCtx(ctx context.Context, username string) (*authdb.UserContext, error) {
	if _, err := c.GetContext(ctx, "admins", username); err != nil {
		return nil, err
	}
	return &authdb.UserContext{
		Name:  username,
		Roles: []string{"_admin"},
	}, nil
}
