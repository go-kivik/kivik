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

package input

import (
	"encoding/json"
	"io"
	"os"
	"strings"

	"github.com/icza/dyno"
	"github.com/spf13/pflag"
	"gopkg.in/yaml.v3"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

type Input struct {
	data string
	file string
	yaml bool
}

func New() *Input {
	return &Input{}
}

func (i *Input) ConfigFlags(pf *pflag.FlagSet) {
	pf.StringVarP(&i.data, "data", "d", "", "JSON document data.")
	pf.StringVarP(&i.file, "data-file", "D", "", "Read document data from the named file. Use - for stdin. Assumed to be JSON, unless the file extension is .yaml or .yml, or the --yaml flag is used.")
	pf.BoolVar(&i.yaml, "yaml", false, "Treat input data as YAML")
}

// jsonReader converts an io.Reader into a json.Marshaler.
type jsonReader struct{ io.Reader }

var _ json.Marshaler = (*jsonReader)(nil)

// MarshalJSON returns the reader's contents. If the reader is also an io.Closer,
// it is closed.
func (r *jsonReader) MarshalJSON() ([]byte, error) {
	if c, ok := r.Reader.(io.Closer); ok {
		defer c.Close() // nolint:errcheck
	}
	buf, err := io.ReadAll(r)
	return buf, errors.Code(errors.ErrIO, err)
}

// jsonObject turns an arbitrary object into a json.Marshaler.
type jsonObject struct {
	i interface{}
}

var _ json.Marshaler = &jsonObject{}

func (o *jsonObject) MarshalJSON() ([]byte, error) {
	return json.Marshal(o.i)
}

// HasInput returns true if some input has been provided.
func (i *Input) HasInput() bool {
	return i.data != "" || i.file != ""
}

// As unmarshals the data input to target.
func (i *Input) As(target interface{}) error {
	j, err := i.JSONData()
	if err != nil {
		return err
	}
	buf, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return json.Unmarshal(buf, target)
}

func (i *Input) JSONData() (json.Marshaler, error) {
	if !i.yaml {
		if i.data != "" {
			return json.RawMessage(i.data), nil
		}
		switch i.file {
		case "-":
			return &jsonReader{os.Stdin}, nil
		case "":
		default:
			if !strings.HasSuffix(i.file, ".yaml") && !strings.HasSuffix(i.file, ".yml") {
				f, err := os.Open(i.file)
				return &jsonReader{f}, errors.Code(errors.ErrNoInput, err)
			}
		}
	}
	if i.data != "" {
		return yaml2json(io.NopCloser(strings.NewReader(i.data)))
	}
	switch i.file {
	case "-":
		return yaml2json(os.Stdin)
	case "":
	default:
		f, err := os.Open(i.file)
		if err != nil {
			return nil, errors.Code(errors.ErrNoInput, err)
		}
		return yaml2json(f)
	}
	return nil, errors.Code(errors.ErrUsage, "no document data provided")
}

func yaml2json(r io.ReadCloser) (json.Marshaler, error) {
	defer r.Close() // nolint:errcheck

	buf, err := io.ReadAll(r)
	if err != nil {
		return nil, errors.Code(errors.ErrIO, err)
	}

	var doc interface{}
	if err := yaml.Unmarshal(buf, &doc); err != nil {
		return nil, errors.Code(errors.ErrData, err)
	}
	return &jsonObject{dyno.ConvertMapI2MapS(doc)}, nil
}

func (i *Input) RawData() (io.ReadCloser, error) {
	if i.data != "" {
		return io.NopCloser(strings.NewReader(i.data)), nil
	}
	switch i.file {
	case "-":
		return os.Stdin, nil
	case "":
	default:
		f, err := os.Open(i.file)
		return f, errors.Code(errors.ErrNoInput, err)
	}
	return nil, errors.Code(errors.ErrUsage, "no attachment data provided")
}
