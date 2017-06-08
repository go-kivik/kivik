package serve

import (
	"context"
	"testing"

	"github.com/flimzy/kivik/authdb"
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
	s.loadConf()
	tests := []validateTest{
		{Name: "Valid", Cookie: "YWRtaW46NThDNTQzN0Y6OnE2cBAuoQKvVBHF2l4PIqKHqDM", Valid: true,
			User: &authdb.UserContext{Name: "admin", Salt: "foo bar baz"}},
		{Name: "WrongSalt", Cookie: "YWRtaW46NThDNTQzN0Y697rnaWCa_rarAm25wbOg3Gm3mqc", Valid: false,
			User: &authdb.UserContext{Name: "admin", Salt: "123"}},
	}
	for _, test := range tests {
		func(test validateTest) {
			t.Run(test.Name, func(t *testing.T) {
				valid, err := s.ValidateCookie(context.Background(), test.User, test.Cookie)
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

type cookieTest struct {
	TestName string
	Input    string
	Name     string
	Created  int64
	Err      string
}

func TestDecodeCookie(t *testing.T) {
	tests := []cookieTest{
		{TestName: "Valid", Input: "YWRtaW46NThDNTQzN0Y697rnaWCa_rarAm25wbOg3Gm3mqc", Name: "admin", Created: 1489322879},
		{TestName: "InvalidBasae64", Input: "bogus base64", Err: "illegal base64 data at input byte 5"},
		{TestName: "Extra colons", Input: "Zm9vOjEyMzQ1OmJhcjpiYXoK", Name: "foo", Created: 74565},
		{TestName: "InvalidTimestamp", Input: "Zm9vOmJhcjpiYXoK", Err: `invalid timestamp: strconv.ParseInt: parsing "bar": invalid syntax`},
	}
	for _, test := range tests {
		func(test cookieTest) {
			t.Run(test.TestName, func(t *testing.T) {
				name, created, err := DecodeCookie(test.Input)
				var errMsg string
				if err != nil {
					errMsg = err.Error()
				}
				if name != test.Name {
					t.Errorf("Name: Expected '%s', got '%s'", test.Name, name)
				}
				if created != test.Created {
					t.Errorf("Created: Expected %d, got %d", test.Created, created)
				}
				if errMsg != test.Err {
					t.Errorf("Error: Expected '%s', got '%s'", test.Err, errMsg)
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
	s.loadConf()
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
