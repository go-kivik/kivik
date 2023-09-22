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

package friendly

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"
)

type friendly struct {
	io.Reader
}

func (friendly) Execute(w io.Writer) error {
	_, err := w.Write([]byte("Friendly!"))
	return err
}

func TestOutput(t *testing.T) {
	type tt struct {
		r   io.Reader
		err string
	}

	tests := testy.NewTable()
	tests.Add("non-friendly output", tt{
		r: strings.NewReader(`{"foo":"bar"}`),
	})
	tests.Add("friendly output", tt{
		r: friendly{},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		buf := &bytes.Buffer{}
		f := format{}
		err := f.Output(buf, tt.r)
		testy.Error(t, tt.err, err)
		if d := testy.DiffText(testy.Snapshot(t), buf); d != nil {
			t.Error(d)
		}
	})
}
