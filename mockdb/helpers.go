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

package mockdb

import (
	"bytes"
	"encoding/json"
	"io"
	"testing"

	"github.com/go-kivik/kivik/v4/driver"
)

// DocumentT calls Document, and passes any error to t.Fatal.
func DocumentT(t *testing.T, i any) *driver.Document {
	t.Helper()
	doc, err := Document(i)
	if err != nil {
		t.Fatal(err)
	}
	return doc
}

// Document converts i, which should be of a supported type (see below), into
// a document which can be passed to ExpectedGet.WillReturn().
//
// i is checked against the following list of types, in order. If no match
// is found, an error is returned. Attachments is not populated by this method.
//
//   - string, []byte, or json.RawMessage (interpreted as a JSON string)
//   - io.Reader (assumes JSON can be read from the stream)
//   - any other JSON-marshalable object
func Document(i any) (*driver.Document, error) {
	buf, err := toJSON(i)
	if err != nil {
		return nil, err
	}
	var meta struct {
		Rev string `json:"_rev"`
	}
	if err := json.Unmarshal(buf, &meta); err != nil {
		return nil, err
	}
	return &driver.Document{
		Rev:         meta.Rev,
		Body:        io.NopCloser(bytes.NewReader(buf)),
		Attachments: nil,
	}, nil
}

func toJSON(i any) ([]byte, error) {
	switch t := i.(type) {
	case string:
		return []byte(t), nil
	case []byte:
		return t, nil
	case json.RawMessage:
		return t, nil
	}
	buf := &bytes.Buffer{}
	err := json.NewEncoder(buf).Encode(i)
	return buf.Bytes(), err
}
