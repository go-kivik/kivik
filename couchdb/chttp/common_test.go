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

package chttp

import (
	"io"
	"regexp"
	"testing"

	"gitlab.com/flimzy/testy"
)

// statusErrorRE is a modified version of testy.StatusError, which handles
// exit statuses as well.
func statusErrorRE(t *testing.T, expected string, status int, actual error) {
	t.Helper()
	var err string
	var actualStatus int
	if actual != nil {
		err = actual.Error()
		actualStatus = testy.StatusCode(actual)
	}
	match, e := regexp.MatchString(expected, err)
	if e != nil {
		t.Fatal(e)
	}
	if !match {
		t.Errorf("Unexpected error: %s (expected %s)", err, expected)
	}
	if status != actualStatus {
		t.Errorf("Unexpected status code: %d (expected %d) [%s]", actualStatus, status, err)
	}
	if actual != nil {
		t.SkipNow()
	}
}

type errReader struct {
	io.Reader
	err error
}

func (r *errReader) Read(p []byte) (int, error) {
	c, err := r.Reader.Read(p)
	if err == io.EOF {
		err = r.err
	}
	return c, err
}

type errCloser struct {
	io.Reader
	err error
}

func (r *errCloser) Close() error {
	return r.err
}
