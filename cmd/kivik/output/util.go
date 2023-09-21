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
	"text/template"
)

// JSONReader marshals i as JSON.
func JSONReader(i interface{}) io.Reader {
	r, w := io.Pipe()
	go func() {
		err := json.NewEncoder(w).Encode(i)
		_ = w.CloseWithError(err)
	}()
	return r
}

// FriendlyOutput produces friendly output.
type FriendlyOutput interface {
	io.Reader
	Execute(io.Writer) error
}

type tmplReader struct {
	io.Reader
	data interface{}
	tmpl *template.Template
}

var _ FriendlyOutput = &tmplReader{}

func TemplateReader(tmpl string, data interface{}, r io.Reader) FriendlyOutput {
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
