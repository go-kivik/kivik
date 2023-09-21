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

package yaml

import (
	"encoding/json"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
)

type format struct{}

var _ output.Format = &format{}

// New returns the yaml formatter.
func New() output.Format {
	return &format{}
}

func (f *format) Output(w io.Writer, r io.Reader) error {
	var obj interface{}
	if err := json.NewDecoder(r).Decode(&obj); err != nil {
		return err
	}
	return yaml.NewEncoder(w).Encode(obj)
}
