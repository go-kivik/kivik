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
	"encoding/json"
	"io"
	"sync"
	"text/template"
)

// delayedJSONReader waits until the first call to Read to call json.Marshal.
// This saves calling json.Marshal if the reader is never read, and also
// prevents the consumption of the input, in cases like an attachment, which
// indirectly marshals an io.Reader.
type delayedJSONReader struct {
	mu sync.Mutex
	r  io.Reader
	i  any
}

func (d *delayedJSONReader) Read(p []byte) (int, error) {
	d.mu.Lock()
	if d.r == nil {
		r, w := io.Pipe()
		go func() {
			err := json.NewEncoder(w).Encode(d.i)
			_ = w.CloseWithError(err)
		}()
		d.r = r
	}
	d.mu.Unlock()
	return d.r.Read(p)
}

// JSONReader marshals i as JSON.
func JSONReader(i any) io.Reader {
	return &delayedJSONReader{i: i}
}

// FriendlyOutput produces friendly output.
type FriendlyOutput interface {
	io.Reader
	Execute(io.Writer) error
}

type tmplReader struct {
	io.Reader
	data any
	tmpl *template.Template
}

var _ FriendlyOutput = &tmplReader{}

// TemplateReader ...
func TemplateReader(tmpl string, data any, r io.Reader) FriendlyOutput {
	return &tmplReader{
		Reader: r,
		data:   data,
		tmpl:   template.Must(template.New("").Parse(tmpl)),
	}
}

func (t *tmplReader) Execute(w io.Writer) error {
	if rc, ok := t.Reader.(io.ReadCloser); ok {
		defer rc.Close() // nolint:errcheck
	}
	return t.tmpl.Execute(w, t.data)
}
