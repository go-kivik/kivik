// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:build !js
// +build !js

package kivikd

import (
	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
	"github.com/go-kivik/kivik/v4/x/kivikd/cookies"
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
