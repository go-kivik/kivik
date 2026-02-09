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
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

var testOptions = map[string]any{"foo": 123}

func parseTime(t *testing.T, str string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, str)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

type errReader string

var _ io.ReadCloser = errReader("")

func (r errReader) Close() error {
	return nil
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, errors.New(string(r))
}

func body(s string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(s))
}

type mockIterator struct {
	NextFunc  func(any) error
	CloseFunc func() error
}

var _ iterator = &mockIterator{}

func (i *mockIterator) Next(ifce any) error {
	return i.NextFunc(ifce)
}

func (i *mockIterator) Close() error {
	return i.CloseFunc()
}
