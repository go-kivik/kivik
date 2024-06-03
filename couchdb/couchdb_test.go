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

package couchdb

import (
	"net/http"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestNewClient(t *testing.T) {
	type ncTest struct {
		name       string
		dsn        string
		options    driver.Options
		expectedUA []string
		status     int
		err        string
	}
	tests := []ncTest{
		{
			name:   "invalid url",
			dsn:    "foo.com/%xxx",
			status: http.StatusBadRequest,
			err:    `parse "?http://foo.com/%xxx"?: invalid URL escape "%xx"`,
		},
		{
			name:       "success",
			dsn:        "http://foo.com/",
			expectedUA: nil,
		},
		{
			name:    "User Agent",
			dsn:     "http://foo.com/",
			options: OptionUserAgent("test/foo"),
			expectedUA: []string{
				"test/foo",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			driver := &couch{}
			opts := test.options
			if opts == nil {
				opts = mock.NilOption
			}
			result, err := driver.NewClient(test.dsn, opts)
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if err != nil {
				return
			}
			client, ok := result.(*client)
			if !ok {
				t.Errorf("Unexpected type returned: %t", result)
			}
			if d := testy.DiffInterface(test.expectedUA, client.Client.UserAgents); d != nil {
				t.Error(d)
			}
		})
	}
	t.Run("custom HTTP client", func(t *testing.T) {
		opts := OptionHTTPClient(&http.Client{Timeout: time.Millisecond})
		driver := &couch{}
		c, err := driver.NewClient("http://example.com/", opts)
		if err != nil {
			t.Fatal(err)
		}
		if c.(*client).Client.Timeout != time.Millisecond {
			t.Error("Unexpected *http.Client returned")
		}
	})
}

func TestDB(t *testing.T) {
	tests := []struct {
		name     string
		client   *client
		dbName   string
		expected *db
		status   int
		err      string
	}{
		{
			name:   "no dbname",
			status: http.StatusBadRequest,
			err:    "kivik: dbName required",
		},
		{
			name:   "no full commit",
			dbName: "foo",
			expected: &db{
				dbName: "foo",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.DB(test.dbName, mock.NilOption)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if err != nil {
				return
			}
			if _, ok := result.(*db); !ok {
				t.Errorf("Unexpected result type: %T", result)
			}
		})
	}
}
