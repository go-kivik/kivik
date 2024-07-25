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

package kivikd

import (
	"testing"

	"github.com/go-kivik/kivik/v4/x/kivikd/authdb"
)

type validateTest struct {
	Name   string
	Cookie string
	User   *authdb.UserContext
	Valid  bool
	Err    string
}

func TestValidateCookie(t *testing.T) {
	s := &Service{}
	if err := s.loadConf(); err != nil {
		t.Fatal(err)
	}
	tests := []validateTest{
		{
			Name: "Valid", Cookie: "YWRtaW46NThDNTQzN0Y6OnE2cBAuoQKvVBHF2l4PIqKHqDM", Valid: true,
			User: &authdb.UserContext{Name: "admin", Salt: "foo bar baz"},
		},
		{
			Name: "WrongSalt", Cookie: "YWRtaW46NThDNTQzN0Y697rnaWCa_rarAm25wbOg3Gm3mqc", Valid: false,
			User: &authdb.UserContext{Name: "admin", Salt: "123"},
		},
	}
	for _, test := range tests {
		func(test validateTest) {
			t.Run(test.Name, func(t *testing.T) {
				valid, err := s.ValidateCookie(test.User, test.Cookie)
				var errMsg string
				if err != nil {
					errMsg = err.Error()
				}
				if errMsg != test.Err {
					t.Errorf("Unexpected error.\nExpected: %s\n  Actual: %s\n", test.Err, errMsg)
				}
				if valid != test.Valid {
					t.Errorf("Unexpected result. Expected %t, Actual %t", test.Valid, valid)
				}
			})
		}(test)
	}
}

type tokenTest struct {
	Name     string
	Salt     string
	Created  int64
	Expected string
	Err      string
}

func TestCreateAuthToken(t *testing.T) {
	s := &Service{}
	if err := s.loadConf(); err != nil {
		t.Fatal(err)
	}
	tests := []tokenTest{
		{Name: "admin", Salt: "foo bar baz", Created: 1489322879, Expected: "YWRtaW46NThDNTQzN0Y6OnE2cBAuoQKvVBHF2l4PIqKHqDM"},
		{Name: "bob", Salt: "0123456789abc", Created: 1489322879, Expected: "Ym9iOjU4QzU0MzdGOihHwWRLS2vekOgsRrH1cEVrk6za"},
	}

	for _, test := range tests {
		func(test tokenTest) {
			t.Run(test.Name, func(t *testing.T) {
				result, err := s.CreateAuthToken(test.Name, test.Salt, test.Created)
				var errMsg string
				if err != nil {
					errMsg = err.Error()
				}
				if errMsg != test.Err {
					t.Errorf("Unexpected error. Expected '%s', got '%s'", test.Err, errMsg)
				}
				if result != test.Expected {
					t.Errorf("Expected: %s\n  Actual: %s\n", test.Expected, result)
				}
			})
		}(test)
	}
}
