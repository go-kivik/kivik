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

package kivik

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestSession(t *testing.T) {
	tests := []struct {
		name     string
		client   driver.Client
		closed   bool
		expected any
		status   int
		err      string
	}{
		{
			name:   "driver doesn't implement Sessioner",
			client: &mock.Client{},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support sessions",
		},
		{
			name: "driver returns error",
			client: &mock.Sessioner{
				SessionFunc: func(context.Context) (*driver.Session, error) {
					return nil, errors.New("session error")
				},
			},
			status: http.StatusInternalServerError,
			err:    "session error",
		},
		{
			name: "good response",
			client: &mock.Sessioner{
				SessionFunc: func(context.Context) (*driver.Session, error) {
					return &driver.Session{
						Name:  "curly",
						Roles: []string{"stooges"},
					}, nil
				},
			},
			expected: &Session{
				Name:  "curly",
				Roles: []string{"stooges"},
			},
		},
		{
			name:   "closed",
			closed: true,
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			client := &Client{
				driverClient: test.client,
				closed:       test.closed,
			}
			session, err := client.Session(context.Background())
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
			if d := testy.DiffInterface(test.expected, session); d != nil {
				t.Error(d)
			}
		})
	}
}
