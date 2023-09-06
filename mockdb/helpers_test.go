package kivikmock

import (
	"encoding/json"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
)

func TestDocument(t *testing.T) {
	type tst struct {
		i        interface{}
		expected *driver.Document
		content  interface{}
		err      string
	}
	tests := testy.NewTable()
	tests.Add("simple doc", tst{
		i:        map[string]string{"foo": "bar"},
		expected: &driver.Document{},
		content:  []byte(`{"foo":"bar"}`),
	})
	tests.Add("Unmarshalable", tst{
		i:   func() {},
		err: "json: unsupported type: func()",
	})
	tests.Add("raw string", tst{
		i:        `{"foo":"bar"}`,
		expected: &driver.Document{},
		content:  []byte(`{"foo":"bar"}`),
	})
	tests.Add("raw bytes", tst{
		i:        []byte(`{"foo":"bar"}`),
		expected: &driver.Document{},
		content:  []byte(`{"foo":"bar"}`),
	})
	tests.Add("json.RawMessage", tst{
		i:        json.RawMessage(`{"foo":"bar"}`),
		expected: &driver.Document{},
		content:  []byte(`{"foo":"bar"}`),
	})
	tests.Add("rev", tst{
		i: `{"_rev":"1-xxx"}`,
		expected: &driver.Document{
			Rev: "1-xxx",
		},
		content: []byte(`{"_rev":"1-xxx"}`),
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := Document(test.i)
		testy.Error(t, test.err, err)
		if d := testy.DiffAsJSON(test.content, result.Body); d != nil {
			t.Errorf("Unexpected content:\n%s\n", d)
		}
		result.Body.Close() // nolint: errcheck
		result.Body = nil
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}
