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

// Package kt provides common utilities for Kivik tests.
package kt

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/cenkalti/backoff/v4"
)

// testFunc is the signature for registered test functions.
type testFunc func(*testing.T, *Context)

var tests = make(map[string]testFunc)

// Register registers a test to be run for the given test suite.
func Register(name string, fn testFunc) {
	tests[name] = fn
}

// RunSubtests executes the registered tests against the client.
func RunSubtests(t *testing.T, c *Context) { //nolint:thelper
	for name, fn := range tests {
		c.Run(t, name, func(t *testing.T) {
			t.Helper()
			fn(t, c)
		})
	}
}

var (
	rnd   *rand.Rand
	rndMU = &sync.Mutex{}
)

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

// TestDBPrefix is used to prefix temporary database names during tests.
const TestDBPrefix = "kivik$"

var invalidDBCharsRE = regexp.MustCompile(`[^a-z0-9_$\(\)+/-]`)

// TestDBName generates a randomized string suitable for a database name for
// testing.
func TestDBName(t *testing.T) string {
	id := strings.ToLower(t.Name())
	id = invalidDBCharsRE.ReplaceAllString(id, "_")
	id = id[strings.Index(id, "/")+1:]
	id = strings.ReplaceAll(id, "/", "_") + "$"
	rndMU.Lock()
	dbname := fmt.Sprintf("%s%s%016x", TestDBPrefix, id, rnd.Int63())
	rndMU.Unlock()
	return dbname
}

const maxRetries = 5

// Retry will try an operation up to maxRetries times, in case of one of the
// following failures. All other failures are returned.
func Retry(fn func() error) error {
	bo := backoff.WithMaxRetries(backoff.NewExponentialBackOff(), maxRetries)
	var i int
	return backoff.Retry(func() error {
		err := fn()
		if err != nil {
			if shouldRetry(err) {
				i++
				return fmt.Errorf("attempt #%d failed: %w", i, err)
			}
			return backoff.Permanent(err)
		}
		return nil
	}, bo)
}

func shouldRetry(err error) bool {
	if err == nil {
		return false
	}
	var statusErr interface {
		error
		HTTPStatus() int
	}
	if errors.As(err, &statusErr) {
		if status := statusErr.HTTPStatus(); status < http.StatusInternalServerError {
			return false
		}
	}
	var errno syscall.Errno
	if errors.As(err, &errno) {
		switch errno {
		case syscall.ECONNRESET, syscall.EPIPE:
			return true
		}
	}
	urlErr := new(url.Error)
	if errors.As(err, &urlErr) {
		// Seems string comparison is necessary in some cases.
		msg := strings.TrimSpace(urlErr.Error())
		return strings.HasSuffix(msg, ": connection reset by peer") || // Observed in Go 1.18
			strings.HasSuffix(msg, ": broken pipe") // Observed in Go 1.19 & 1.17
	}
	return false
}

// Body turns a string into a read closer, useful as a request or attachment
// body.
func Body(str string, args ...any) io.ReadCloser {
	return io.NopCloser(strings.NewReader(fmt.Sprintf(str, args...)))
}
