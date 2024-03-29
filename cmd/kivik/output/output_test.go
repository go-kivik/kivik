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

package output

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func Test_ensurenewlineEnding(t *testing.T) {
	t.Run("no newline", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := ensureNewlineEnding(buf)
		_, _ = io.Copy(w, strings.NewReader("asdf"))
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if buf.String() != "asdf\n" {
			t.Errorf("Unexpected output: %q", buf.String())
		}
	})
	t.Run("with newline", func(t *testing.T) {
		buf := &bytes.Buffer{}
		w := ensureNewlineEnding(buf)
		_, _ = io.Copy(w, strings.NewReader("asdf\n"))
		if err := w.Close(); err != nil {
			t.Fatal(err)
		}
		if buf.String() != "asdf\n" {
			t.Errorf("Unexpected output: %q", buf.String())
		}
	})
}

func TestJSONReader(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		in := map[string]interface{}{
			"foo": "bar",
		}
		r := JSONReader(in)
		got, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		want := `{"foo":"bar"}
`
		if d := cmp.Diff(want, string(got)); d != "" {
			t.Error(d)
		}
	})
	t.Run("error", func(t *testing.T) {
		in := make(chan int)
		r := JSONReader(in)
		_, err := io.ReadAll(r)
		want := "json: unsupported type: chan int"
		if err == nil || err.Error() != want {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}
