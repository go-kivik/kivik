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

package couchdb

import (
	"encoding/json"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestEncodeKey(t *testing.T) {
	type tst struct {
		input    interface{}
		expected string
		status   int
		err      string
	}
	tests := testy.NewTable()
	tests.Add("string", tst{
		input:    "foo",
		expected: `"foo"`,
	})
	tests.Add("chan", tst{
		input:  make(chan int),
		status: http.StatusBadRequest,
		err:    "json: unsupported type: chan int",
	})
	tests.Add("[]byte", tst{
		input:    []byte("foo"),
		expected: `"Zm9v"`,
	})
	tests.Add("json.RawMessage", tst{
		input:    json.RawMessage(`"foo"`),
		expected: `"foo"`,
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := encodeKey(test.input)
		testy.StatusError(t, test.err, test.status, err)
		if d := testy.DiffJSON([]byte(test.expected), []byte(result)); d != nil {
			t.Error(d)
		}
	})
}

func TestEncodeKeys(t *testing.T) {
	type tst struct {
		input    map[string]interface{}
		expected map[string]interface{}
		status   int
		err      string
	}
	type keyStruct struct {
		Foo string
		Bar interface{} `json:",omitempty"`
	}
	tests := testy.NewTable()
	tests.Add("nil", tst{
		input:    nil,
		expected: nil,
	})
	tests.Add("unmarshalable", tst{
		input: map[string]interface{}{
			"key": make(chan int),
		},
		status: http.StatusBadRequest,
		err:    "json: unsupported type: chan int",
	})
	tests.Add("unaltered", tst{
		input: map[string]interface{}{
			"foo": 123,
		},
		expected: map[string]interface{}{
			"foo": 123,
		},
	})
	tests.Add("key", tst{
		input: map[string]interface{}{
			"key": 123,
		},
		expected: map[string]interface{}{
			"key": "123",
		},
	})
	tests.Add("keys []interface{}", tst{
		input: map[string]interface{}{
			"foo":  123,
			"keys": []interface{}{"foo", 123},
		},
		expected: map[string]interface{}{
			"foo":  123,
			"keys": `["foo",123]`,
		},
	})
	tests.Add("keys []interface{} invalid", tst{
		input: map[string]interface{}{
			"foo":  123,
			"keys": []interface{}{"foo", 123, make(chan int)},
		},
		status: http.StatusBadRequest,
		err:    "json: unsupported type: chan int",
	})
	tests.Add("keys string", tst{
		input: map[string]interface{}{
			"foo":  123,
			"keys": []string{"foo", "123"},
		},
		expected: map[string]interface{}{
			"foo":  123,
			"keys": `["foo","123"]`,
		},
	})
	tests.Add("keys structs", tst{
		input: map[string]interface{}{
			"keys": []keyStruct{
				{Foo: "abc"},
				{Foo: "xyz"},
			},
		},
		expected: map[string]interface{}{
			"keys": `[{"Foo":"abc"},{"Foo":"xyz"}]`,
		},
	})
	tests.Add("keys structs invalid", tst{
		input: map[string]interface{}{
			"keys": []keyStruct{
				{Foo: "abc", Bar: make(chan int)},
				{Foo: "xyz"},
			},
		},
		status: http.StatusBadRequest,
		err:    "json: unsupported type: chan int",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		err := encodeKeys(test.input)
		testy.StatusError(t, test.err, test.status, err)
		if d := testy.DiffInterface(test.expected, test.input); d != nil {
			t.Error(d)
		}
	})
}
