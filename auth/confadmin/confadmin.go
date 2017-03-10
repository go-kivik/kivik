// Package confadmin provides an authentication service for admins configured
// in server configuration.
package confadmin

import (
	"context"
	"crypto/sha1"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/pbkdf2"

	"github.com/flimzy/kivik/auth"
	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/errors"
)

type conf struct {
	*config.Config
}

var _ auth.Handler = &conf{}

// New returns a new confadmin authentication service provider.
func New(c *config.Config) auth.Handler {
	return &conf{c}
}

func (c *conf) Validate(ctx context.Context, username, password string) (bool, error) {
	hash, err := c.GetContext(ctx, "admins", username)
	if err != nil {
		if errors.StatusCode(err) == errors.StatusNotFound {
			return false, nil
		}
		return false, err
	}
	derivedKey, salt, iterations, err := keySaltIter(hash)
	if err != nil {
		return false, errors.Wrap(err, "unrecognized password hash")
	}
	key := fmt.Sprintf("%x", pbkdf2.Key([]byte(password), []byte(salt), iterations, auth.PBKDF2KeyLength, sha1.New))
	return key == derivedKey, nil
}

const hashPrefix = "-" + auth.SchemePBKDF2 + "-"

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

func (c *conf) Roles(ctx context.Context, username string) ([]string, error) {
	if _, err := c.GetContext(ctx, "admins", username); err != nil {
		return nil, err
	}
	return []string{"_admin"}, nil
}
