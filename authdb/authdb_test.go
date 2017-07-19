package authdb

import "testing"

func TestCreateAuthToken(t *testing.T) {
	type catTest struct {
		name                   string
		username, salt, secret string
		time                   int64
		expected               string
		recovery               string
	}
	tests := []catTest{
		{
			name:     "no secret",
			recovery: "secret must be set",
		},
		{
			name:     "no salt",
			secret:   "foo",
			recovery: "salt must be set",
		},
		{
			name:     "valid",
			secret:   "foo",
			salt:     "4e170ffeb6f34daecfd814dfb4001a73",
			username: "baz",
			time:     12345,
			expected: "YmF6OjMwMzk6fqvvtjD0eY7M_OZCiSWeRk01mlo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recovery := func() (recovery string) {
				defer func() {
					if r := recover(); r != nil {
						recovery = r.(string)
					}
				}()
				result := CreateAuthToken(test.username, test.salt, test.secret, test.time)
				if result != test.expected {
					t.Errorf("Unexpected result: %s", result)
				}
				return ""
			}()
			if recovery != test.recovery {
				t.Errorf("Unexpected panic recovery: %s", recovery)
			}
		})
	}
}
