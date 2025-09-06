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

//go:build !js

package options

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestViewOptions(t *testing.T) {
	type test struct {
		options    driver.Options
		view       string
		want       *ViewOptions
		wantErr    string
		wantStatus int
	}

	tests := testy.NewTable()
	tests.Add("all defaults", test{
		options: mock.NilOption,
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("conflicts: invalid string", test{
		options:    kivik.Param("conflicts", "oink"),
		wantErr:    "invalid value for 'conflicts': oink",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("conflicts: string", test{
		options: kivik.Param("conflicts", "true"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			conflicts:    true,
		},
	})
	tests.Add("conflicts: bool", test{
		options: kivik.Param("conflicts", true),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			conflicts:    true,
		},
	})
	tests.Add("conflicts: int", test{
		options:    kivik.Param("conflicts", 3),
		wantErr:    "invalid value for 'conflicts': 3",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("descending: bool", test{
		options: kivik.Param("descending", true),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			descending:   true,
		},
	})
	tests.Add("descending: string", test{
		options: kivik.Param("descending", "true"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			descending:   true,
		},
	})
	tests.Add("descending: invalid string", test{
		options:    kivik.Param("descending", "chicken"),
		wantErr:    "invalid value for 'descending': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("descending: int", test{
		options:    kivik.Param("descending", 3),
		wantErr:    "invalid value for 'descending': 3",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("endkey: invalid json", test{
		options:    kivik.Param("endkey", json.RawMessage("oink")),
		wantErr:    `invalid value for 'endkey': invalid character 'o' looking for beginning of value in key`,
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("endkey: valid json", test{
		options: kivik.Param("endkey", json.RawMessage(`"oink"`)),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			endkey:       `"oink"`,
		},
	})
	tests.Add("endkey: plain string", test{
		options: kivik.Param("endkey", "oink"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			endkey:       `"oink"`,
		},
	})
	tests.Add("endkey: unmarshalable value", test{
		options:    kivik.Param("endkey", func() {}),
		wantErr:    `invalid value for 'endkey': json: unsupported type: func() in key`,
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("endkey: slice", test{
		options: kivik.Param("endkey", []string{"foo", "bar"}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			endkey:       `["foo","bar"]`,
		},
	})
	tests.Add("end_key: invalid json", test{
		options:    kivik.Param("end_key", json.RawMessage("oink")),
		wantErr:    `invalid value for 'end_key': invalid character 'o' looking for beginning of value in key`,
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("end_key: valid json", test{
		options: kivik.Param("end_key", json.RawMessage(`"oink"`)),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			endkey:       `"oink"`,
		},
	})
	tests.Add("startkey: invalid json", test{
		options:    kivik.Param("startkey", json.RawMessage("oink")),
		wantErr:    `invalid value for 'startkey': invalid character 'o' looking for beginning of value in key`,
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("startkey: valid json", test{
		options: kivik.Param("startkey", json.RawMessage(`"oink"`)),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			startkey:     `"oink"`,
		},
	})
	tests.Add("start_key: invalid json", test{
		options:    kivik.Param("start_key", json.RawMessage("oink")),
		wantErr:    `invalid value for 'start_key': invalid character 'o' looking for beginning of value in key`,
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("start_key: valid json", test{
		options: kivik.Param("start_key", json.RawMessage(`"oink"`)),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			startkey:     `"oink"`,
		},
	})
	tests.Add("group: bool", test{
		options: kivik.Param("group", true),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
			sorted:       true,
			group:        true,
		},
	})
	tests.Add("group: valid string", test{
		options: kivik.Param("group", "true"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
			sorted:       true,
			group:        true,
		},
	})
	tests.Add("group: invalid string", test{
		options:    kivik.Param("group", "chicken"),
		wantErr:    "invalid value for 'group': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("group: int", test{
		options:    kivik.Param("group", 3),
		wantErr:    "invalid value for 'group': 3",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("group_level: int", test{
		options: kivik.Param("group_level", 3),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
			sorted:       true,
			group:        true,
			groupLevel:   3,
		},
	})
	tests.Add("group_level: valid string", test{
		options: kivik.Param("group_level", "3"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
			sorted:       true,
			group:        true,
			groupLevel:   3,
		},
	})
	tests.Add("group_level: invalid string", test{
		options:    kivik.Param("group_level", "chicken"),
		wantErr:    "invalid value for 'group_level': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("group_level: float64", test{
		options: kivik.Param("group_level", 3.0),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
			sorted:       true,
			group:        true,
			groupLevel:   3,
		},
	})
	tests.Add("include_docs: bool", test{
		options: kivik.Param("include_docs", true),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			includeDocs:  true,
		},
	})
	tests.Add("include_docs: valid string", test{
		options: kivik.Param("include_docs", "true"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			includeDocs:  true,
		},
	})
	tests.Add("include_docs: invalid string", test{
		options:    kivik.Param("include_docs", "chicken"),
		wantErr:    "invalid value for 'include_docs': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("include_docs: int", test{
		options:    kivik.Param("include_docs", 3),
		wantErr:    "invalid value for 'include_docs': 3",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("attachments: bool", test{
		options: kivik.Param("attachments", true),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			attachments:  true,
		},
	})
	tests.Add("attachments: valid string", test{
		options: kivik.Param("attachments", "true"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			attachments:  true,
		},
	})
	tests.Add("attachments: invalid string", test{
		options:    kivik.Param("attachments", "chicken"),
		wantErr:    "invalid value for 'attachments': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("attachments: int", test{
		options:    kivik.Param("attachments", 3),
		wantErr:    "invalid value for 'attachments': 3",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("inclusive_end: bool", test{
		options: kivik.Param("inclusive_end", false),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: false,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("inclusive_end: valid string", test{
		options: kivik.Param("inclusive_end", "false"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: false,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("inclusive_end: invalid string", test{
		options:    kivik.Param("inclusive_end", "chicken"),
		wantErr:    "invalid value for 'inclusive_end': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("inclusive_end: int", test{
		options:    kivik.Param("inclusive_end", 3),
		wantErr:    "invalid value for 'inclusive_end': 3",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("limit: int", test{
		options: kivik.Param("limit", 3),
		want: &ViewOptions{
			limit:        3,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("limit: float64", test{
		options: kivik.Param("limit", 3.0),
		want: &ViewOptions{
			limit:        3,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("limit: valid string", test{
		options: kivik.Param("limit", "3"),
		want: &ViewOptions{
			limit:        3,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("limit: valid string2", test{
		options: kivik.Param("limit", "3.0"),
		want: &ViewOptions{
			limit:        3,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("limit: invalid string", test{
		options:    kivik.Param("limit", "chicken"),
		wantErr:    "invalid value for 'limit': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("limit: slice", test{
		options:    kivik.Param("limit", []int{1, 2}),
		wantErr:    "invalid value for 'limit': [1 2]",
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("skip: int", test{
		options: kivik.Param("skip", 3),
		want: &ViewOptions{
			limit:        -1,
			skip:         3,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("skip: float64", test{
		options: kivik.Param("skip", 3.0),
		want: &ViewOptions{
			limit:        -1,
			skip:         3,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("skip: valid string", test{
		options: kivik.Param("skip", "3"),
		want: &ViewOptions{
			limit:        -1,
			skip:         3,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("skip: valid string2", test{
		options: kivik.Param("skip", "3.0"),
		want: &ViewOptions{
			limit:        -1,
			skip:         3,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("skip: invalid string", test{
		options:    kivik.Param("skip", "chicken"),
		wantErr:    "invalid value for 'skip': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("skip: slice", test{
		options:    kivik.Param("skip", []int{1, 2}),
		wantErr:    "invalid value for 'skip': [1 2]",
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("reduce: bool", test{
		options: kivik.Param("reduce", true),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			reduce:       &[]bool{true}[0],
		},
	})
	tests.Add("reduce: valid string", test{
		options: kivik.Param("reduce", "true"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			reduce:       &[]bool{true}[0],
		},
	})
	tests.Add("reduce: invalid string", test{
		options:    kivik.Param("reduce", "chicken"),
		wantErr:    "invalid value for 'reduce': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("reduce: int", test{
		options:    kivik.Param("reduce", 3),
		wantErr:    "invalid value for 'reduce': 3",
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("update: bool", test{
		options: kivik.Param("update", true),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("update: valid string", test{
		options: kivik.Param("update", "true"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
		},
	})
	tests.Add("update: valid string2", test{
		options: kivik.Param("update", "lazy"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "lazy",
			sorted:       true,
		},
	})
	tests.Add("update: invalid string", test{
		options:    kivik.Param("update", "chicken"),
		wantErr:    "invalid value for 'update': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("update: int", test{
		options:    kivik.Param("update", 3),
		wantErr:    "invalid value for 'update': 3",
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("update_seq: bool", test{
		options: kivik.Param("update_seq", true),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			updateSeq:    true,
		},
	})
	tests.Add("update_seq: valid string", test{
		options: kivik.Param("update_seq", "true"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			updateSeq:    true,
		},
	})
	tests.Add("update_seq: invalid string", test{
		options:    kivik.Param("update_seq", "chicken"),
		wantErr:    "invalid value for 'update_seq': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("update_seq: int", test{
		options:    kivik.Param("update_seq", 3),
		wantErr:    "invalid value for 'update_seq': 3",
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("endkey_docid: string", test{
		options: kivik.Param("endkey_docid", "oink"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			endkeyDocID:  `oink`,
		},
	})
	tests.Add("endkey_docid: raw json", test{
		options:    kivik.Param("endkey_docid", json.RawMessage(`"oink"`)),
		wantErr:    `invalid value for 'endkey_docid': [34 111 105 110 107 34]`,
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("end_key_doc_id: string", test{
		options: kivik.Param("end_key_doc_id", "oink"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			endkeyDocID:  `oink`,
		},
	})
	tests.Add("end_key_doc_id: raw json", test{
		options:    kivik.Param("end_key_doc_id", json.RawMessage(`"oink"`)),
		wantErr:    `invalid value for 'end_key_doc_id': [34 111 105 110 107 34]`,
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("key: valid string", test{
		options: kivik.Param("key", "oink"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			key:          `"oink"`,
		},
	})
	tests.Add("key: valid json", test{
		options: kivik.Param("key", json.RawMessage(`"oink"`)),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			key:          `"oink"`,
		},
	})
	tests.Add("key: invalid json", test{
		options:    kivik.Param("key", json.RawMessage(`oink`)),
		wantErr:    `invalid value for 'key': invalid character 'o' looking for beginning of value in key`,
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("keys: slice of strings", test{
		options: kivik.Param("keys", []string{"foo", "bar"}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			keys:         []string{`"foo"`, `"bar"`},
		},
	})
	tests.Add("keys: slice of ints", test{
		options: kivik.Param("keys", []int{1, 2, 3}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			keys:         []string{`1`, `2`, `3`},
		},
	})
	tests.Add("keys: slice of mixed values", test{
		options: kivik.Param("keys", []any{"one", 2, [3]int{1, 2, 3}, json.RawMessage(`"foo"`)}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			keys:         []string{`"one"`, `2`, `[1,2,3]`, `"foo"`},
		},
	})
	tests.Add("keys: valid raw JSON", test{
		options: kivik.Param("keys", json.RawMessage(`["foo", "bar"]`)),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			keys:         []string{`"foo"`, `"bar"`},
		},
	})
	tests.Add("keys: invalid raw JSON", test{
		options:    kivik.Param("keys", json.RawMessage(`invalid`)),
		wantErr:    `invalid value for 'keys': invalid character 'i' looking for beginning of value`,
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("keys: unmarshalable type", test{
		options:    kivik.Param("keys", []any{func() {}}),
		wantErr:    `invalid value for 'keys': json: unsupported type: func()`,
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("keys: non slice/array type", test{
		options:    kivik.Param("keys", "foo"),
		wantErr:    `invalid value for 'keys': json: cannot unmarshal string into Go value of type []json.RawMessage`,
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("startkey_docid: string", test{
		options: kivik.Param("startkey_docid", "oink"),
		want: &ViewOptions{
			limit:         -1,
			inclusiveEnd:  true,
			update:        "true",
			sorted:        true,
			startkeyDocID: `oink`,
		},
	})
	tests.Add("startkey_docid: raw json", test{
		options:    kivik.Param("startkey_docid", json.RawMessage(`"oink"`)),
		wantErr:    `invalid value for 'startkey_docid': [34 111 105 110 107 34]`,
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("start_key_doc_id: string", test{
		options: kivik.Param("start_key_doc_id", "oink"),
		want: &ViewOptions{
			limit:         -1,
			inclusiveEnd:  true,
			update:        "true",
			sorted:        true,
			startkeyDocID: `oink`,
		},
	})
	tests.Add("start_key_doc_id: raw json", test{
		options:    kivik.Param("start_key_doc_id", json.RawMessage(`"oink"`)),
		wantErr:    `invalid value for 'start_key_doc_id': [34 111 105 110 107 34]`,
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("sorted: bool", test{
		options: kivik.Param("sorted", false),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       false,
		},
	})
	tests.Add("sorted: false, but re-enabled by descending=true", test{
		options: kivik.Params(map[string]any{
			"sorted":     false,
			"descending": true,
		}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			descending:   true,
			sorted:       true,
		},
	})
	tests.Add("sorted: false, but re-enabled by descending=false", test{
		options: kivik.Params(map[string]any{
			"sorted":     false,
			"descending": false,
		}), want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			descending:   false,
			sorted:       true,
		},
	})
	tests.Add("sorted: valid string", test{
		options: kivik.Param("sorted", "false"),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       false,
		},
	})
	tests.Add("sorted: invalid string", test{
		options:    kivik.Param("sorted", "chicken"),
		wantErr:    "invalid value for 'sorted': chicken",
		wantStatus: http.StatusBadRequest,
	})

	tests.Add("att_encoding_info: bool", test{
		options: kivik.Param("att_encoding_info", true),
		want: &ViewOptions{
			limit:           -1,
			inclusiveEnd:    true,
			update:          "true",
			sorted:          true,
			attEncodingInfo: true,
		},
	})
	tests.Add("att_encoding_info: valid string", test{
		options: kivik.Param("att_encoding_info", "true"),
		want: &ViewOptions{
			limit:           -1,
			inclusiveEnd:    true,
			update:          "true",
			sorted:          true,
			attEncodingInfo: true,
		},
	})
	tests.Add("att_encoding_info: invalid string", test{
		options:    kivik.Param("att_encoding_info", "chicken"),
		wantErr:    "invalid value for 'att_encoding_info': chicken",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("keys + key", test{
		options: kivik.Params(map[string]any{
			"key":  "a",
			"keys": []string{"b", "c"},
		}),
		wantErr:    "`keys` is incompatible with `key`, `start_key` and `end_key`",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("keys + endkey", test{
		options: kivik.Params(map[string]any{
			"endkey": "a",
			"keys":   []string{"b", "c"},
		}),
		wantErr:    "`keys` is incompatible with `key`, `start_key` and `end_key`",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("keys + startkey", test{
		options: kivik.Params(map[string]any{
			"startkey": "a",
			"keys":     []string{"b", "c"},
		}),
		wantErr:    "`keys` is incompatible with `key`, `start_key` and `end_key`",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("key & startkey conflict", test{
		options: kivik.Params(map[string]any{
			"startkey": "d",
			"key":      "a",
		}),
		wantErr:    "no rows can match your key range, change your start_key or key or set descending=true",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("key & startkey no conflict", test{
		options: kivik.Params(map[string]any{
			"startkey": "a",
			"key":      "a",
		}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			startkey:     `"a"`,
			key:          `"a"`,
		},
	})
	tests.Add("key & startkey + descending", test{
		options: kivik.Params(map[string]any{
			"startkey":   "d",
			"key":        "a",
			"descending": true,
		}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			descending:   true,
			startkey:     `"d"`,
			key:          `"a"`,
		},
	})
	tests.Add("key & startkey + descending conflict", test{
		options: kivik.Params(map[string]any{
			"startkey":   "a",
			"key":        "b",
			"descending": true,
		}),
		wantErr:    "no rows can match your key range, change your start_key or key or set descending=false",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("key & endkey conflict", test{
		options: kivik.Params(map[string]any{
			"endkey": "a",
			"key":    "b",
		}),
		wantErr:    "no rows can match your key range, reverse your end_key or key or set descending=true",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("key & endkey no conflict", test{
		options: kivik.Params(map[string]any{
			"endkey": "a",
			"key":    "a",
		}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			endkey:       `"a"`,
			key:          `"a"`,
		},
	})
	tests.Add("key & endkey + descending", test{
		options: kivik.Params(map[string]any{
			"endkey":     "a",
			"key":        "b",
			"descending": true,
		}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			descending:   true,
			endkey:       `"a"`,
			key:          `"b"`,
		},
	})
	tests.Add("key & endkey + descending conflict", test{
		options: kivik.Params(map[string]any{
			"endkey":     "b",
			"key":        "a",
			"descending": true,
		}),
		wantErr:    "no rows can match your key range, reverse your end_key or key or set descending=false",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("endkey and startkey conflict", test{
		options: kivik.Params(map[string]any{
			"startkey": "z",
			"endkey":   "a",
		}),
		wantErr:    "no rows can match your key range, reverse your start_key and end_key or set descending=true",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("endkey and startkey no conflict", test{
		options: kivik.Params(map[string]any{
			"startkey": "a",
			"endkey":   "z",
		}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			startkey:     `"a"`,
			endkey:       `"z"`,
		},
	})
	tests.Add("endkey and startkey + descending conflict", test{
		options: kivik.Params(map[string]any{
			"startkey":   "a",
			"endkey":     "z",
			"descending": true,
		}),
		wantErr:    "no rows can match your key range, reverse your start_key and end_key or set descending=false",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("endkey and startkey + descending no conflict", test{
		options: kivik.Params(map[string]any{
			"startkey":   "z",
			"endkey":     "a",
			"descending": true,
		}),
		want: &ViewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			sorted:       true,
			descending:   true,
			startkey:     `"z"`,
			endkey:       `"a"`,
		},
	})
	tests.Add("key falls outof range of startkey-endkey", test{
		options: kivik.Params(map[string]any{
			"startkey": "a",
			"endkey":   "q",
			"key":      "z",
		}),
		wantErr:    "no rows can match your key range, change your start_key, end_key, or key",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("key falls outof range of startkey-endkey even when descending", test{
		options: kivik.Params(map[string]any{
			"startkey":   "a",
			"endkey":     "q",
			"key":        "z",
			"descending": true,
		}),
		wantErr:    "no rows can match your key range, reverse your start_key and end_key or set descending=false",
		wantStatus: http.StatusBadRequest,
	})
	/*
		stable (boolean) – Whether or not the view results should be returned from a stable set of shards. Default is false.
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		got, err := New(tt.options).ViewOptions(tt.view)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("unexpected error: %v", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		if d := cmp.Diff(tt.want, got, cmp.AllowUnexported(ViewOptions{})); d != "" {
			t.Errorf("Unexpected result:\n%s", d)
		}
	})
}
