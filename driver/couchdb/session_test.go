package couchdb

import (
	"context"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik/driver"
)

func TestSession(t *testing.T) {
	tests := []struct {
		name     string
		client   *client
		expected interface{}
		err      string
	}{
		{
			name:   "valid",
			client: getClient(t),
			expected: &driver.Session{
				Name:                   "admin",
				Roles:                  []string{"_admin"},
				AuthenticationMethod:   "cookie",
				AuthenticationHandlers: []string{"oauth", "cookie", "default"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			session, err := test.client.Session(context.Background())
			var errMsg string
			if err != nil {
				errMsg = err.Error()
			}
			if errMsg != test.err {
				t.Errorf("Unexpected error: %s", errMsg)
			}
			if err != nil {
				return
			}
			session.RawResponse = nil // For consistent check
			if d := diff.Interface(test.expected, session); d != nil {
				t.Error(d)
			}
		})
	}
}
