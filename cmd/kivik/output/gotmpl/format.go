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

package gotmpl

import (
	"encoding/json"
	"io"
	"text/template"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
)

type format struct {
	tmpl *template.Template
}

var _ output.Format = &format{}

// New returns a go-template formatter.
func New() output.Format {
	return &format{}
}

func (format) Required() bool { return true }

func (f *format) Arg(arg string) error {
	var err error
	f.tmpl, err = template.New("").Parse(arg)
	return err
}

func (f *format) Output(w io.Writer, r io.Reader) error {
	var obj interface{}
	if err := json.NewDecoder(r).Decode(&obj); err != nil {
		return err
	}
	return f.tmpl.Execute(w, obj)
}
