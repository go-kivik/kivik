package kivikd

import (
	"github.com/go-kivik/kivikd/v4/authdb"
	"github.com/go-kivik/kivikd/v4/cookies"
)

// CreateAuthToken hashes a user name, salt, timestamp, and the server secret
// into an authentication token.
func (s *Service) CreateAuthToken(name, salt string, time int64) (string, error) {
	secret := s.getAuthSecret()
	return authdb.CreateAuthToken(name, salt, secret, time), nil
}

// ValidateCookie validates a cookie against a user context.
func (s *Service) ValidateCookie(user *authdb.UserContext, cookie string) (bool, error) {
	name, t, err := cookies.DecodeCookie(cookie)
	if err != nil {
		return false, err
	}
	token, err := s.CreateAuthToken(name, user.Salt, t)
	if err != nil {
		return false, err
	}
	return token == cookie, nil
}
