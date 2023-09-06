package kivikmock

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"testing"

	"github.com/go-kivik/kivik/v4/driver"
)

// DocumentT calls Document, and passes any error to t.Fatal.
func DocumentT(t *testing.T, i interface{}) *driver.Document {
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
func Document(i interface{}) (*driver.Document, error) {
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
		Body:        ioutil.NopCloser(bytes.NewReader(buf)),
		Attachments: nil,
	}, nil
}

func toJSON(i interface{}) ([]byte, error) {
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
