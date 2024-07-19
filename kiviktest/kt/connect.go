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

package kt

import (
	"net/url"
	"os"
	"testing"

	kivik "github.com/go-kivik/kivik/v4"
)

// DSN3 returns a testing DSN from the environment for CouchDB 3.x.
func DSN3(t *testing.T) string {
	t.Helper()
	for _, env := range []string{
		"KIVIK_TEST_DSN_COUCH33",
		"KIVIK_TEST_DSN_COUCH32",
		"KIVIK_TEST_DSN_COUCH31",
		"KIVIK_TEST_DSN_COUCH30",
	} {
		dsn := os.Getenv(env)
		if dsn != "" {
			return dsn
		}
	}
	t.Skip("DSN not set")
	return ""
}

// NoAuthDSN returns a testing DSN with credentials stripped.
func NoAuthDSN(t *testing.T) string {
	t.Helper()
	dsn := DSN3(t)
	parsed, err := url.Parse(dsn)
	if err != nil {
		t.Fatalf("invalid DSN: %s", err)
	}
	parsed.User = nil
	return parsed.String()
}

func connect(dsn string, t *testing.T) *kivik.Client {
	t.Helper()
	client, err := kivik.New("couch", dsn)
	if err != nil {
		t.Fatalf("Failed to connect to '%s': %s", dsn, err)
	}
	return client
}

// GetClient returns a connection to a CouchDB client, for testing.
func GetClient(t *testing.T) *kivik.Client {
	t.Helper()
	return connect(DSN3(t), t)
}

// GetNoAuthClient returns an unauthenticated connection to a CouchDB client, for testing.
func GetNoAuthClient(t *testing.T) *kivik.Client {
	t.Helper()
	return connect(NoAuthDSN(t), t)
}
