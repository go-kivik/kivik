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

// Package json produces JSON output.
package json

import (
	"encoding/json"
	"io"

	"github.com/go-kivik/kivik/v4/cmd/kivik/output"
)

type format struct {
	indent string
}

var (
	_ output.Format    = &format{}
	_ output.FormatArg = &format{}
)

// New returns a json formatter.
func New() output.Format {
	return &format{
		indent: "\t",
	}
}

func (format) Required() bool { return false }

func (f *format) Arg(arg string) error {
	f.indent = arg
	return nil
}

func (f *format) Output(w io.Writer, r io.Reader) error {
	var obj interface{}
	if err := json.NewDecoder(r).Decode(&obj); err != nil {
		return err
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", f.indent)
	return enc.Encode(obj)
}
