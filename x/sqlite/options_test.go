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

package sqlite

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/x/sqlite/v4/internal/mock"
)

func Test_viewOptions(t *testing.T) {
	type test struct {
		options    driver.Options
		view       string
		want       *viewOptions
		wantErr    string
		wantStatus int
	}

	tests := testy.NewTable()
	tests.Add("all defaults", test{
		options: mock.NilOption,
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
		},
	})
	tests.Add("confclits: invalid string", test{
		options:    kivik.Param("conflicts", "oink"),
		wantErr:    "invalid value for 'conflicts': oink",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("conflicts: string", test{
		options: kivik.Param("conflicts", "true"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			conflicts:    true,
		},
	})
	tests.Add("conflicts: bool", test{
		options: kivik.Param("conflicts", true),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			descending:   true,
		},
	})
	tests.Add("descending: string", test{
		options: kivik.Param("descending", "true"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			endkey:       `"oink"`,
		},
	})
	tests.Add("endkey: plain string", test{
		options: kivik.Param("endkey", "oink"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			startkey:     `"oink"`,
		},
	})
	tests.Add("group: bool", test{
		options: kivik.Param("group", true),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
			group:        true,
		},
	})
	tests.Add("group: valid string", test{
		options: kivik.Param("group", "true"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
			group:        true,
			groupLevel:   3,
		},
	})
	tests.Add("group_level: valid string", test{
		options: kivik.Param("group_level", "3"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			reduce:       &[]bool{true}[0],
			update:       "true",
			group:        true,
			groupLevel:   3,
		},
	})
	tests.Add("include_docs: bool", test{
		options: kivik.Param("include_docs", true),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			includeDocs:  true,
		},
	})
	tests.Add("include_docs: valid string", test{
		options: kivik.Param("include_docs", "true"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			attachments:  true,
		},
	})
	tests.Add("attachments: valid string", test{
		options: kivik.Param("attachments", "true"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: false,
			update:       "true",
		},
	})
	tests.Add("inclusive_end: valid string", test{
		options: kivik.Param("inclusive_end", "false"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: false,
			update:       "true",
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
		want: &viewOptions{
			limit:        3,
			inclusiveEnd: true,
			update:       "true",
		},
	})
	tests.Add("limit: float64", test{
		options: kivik.Param("limit", 3.0),
		want: &viewOptions{
			limit:        3,
			inclusiveEnd: true,
			update:       "true",
		},
	})
	tests.Add("limit: valid string", test{
		options: kivik.Param("limit", "3"),
		want: &viewOptions{
			limit:        3,
			inclusiveEnd: true,
			update:       "true",
		},
	})
	tests.Add("limit: valid string2", test{
		options: kivik.Param("limit", "3.0"),
		want: &viewOptions{
			limit:        3,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			skip:         3,
			inclusiveEnd: true,
			update:       "true",
		},
	})
	tests.Add("skip: float64", test{
		options: kivik.Param("skip", 3.0),
		want: &viewOptions{
			limit:        -1,
			skip:         3,
			inclusiveEnd: true,
			update:       "true",
		},
	})
	tests.Add("skip: valid string", test{
		options: kivik.Param("skip", "3"),
		want: &viewOptions{
			limit:        -1,
			skip:         3,
			inclusiveEnd: true,
			update:       "true",
		},
	})
	tests.Add("skip: valid string2", test{
		options: kivik.Param("skip", "3.0"),
		want: &viewOptions{
			limit:        -1,
			skip:         3,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			reduce:       &[]bool{true}[0],
		},
	})
	tests.Add("reduce: valid string", test{
		options: kivik.Param("reduce", "true"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
		},
	})
	tests.Add("update: valid string", test{
		options: kivik.Param("update", "true"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
		},
	})
	tests.Add("update: valid string2", test{
		options: kivik.Param("update", "lazy"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "lazy",
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
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
			updateSeq:    true,
		},
	})
	tests.Add("update_seq: valid string", test{
		options: kivik.Param("update_seq", "true"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			update:       "true",
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

	/*
		endkey_docid (string) – Stop returning records when the specified document ID is reached. Ignored if endkey is not set.

		end_key_doc_id (string) – Alias for endkey_docid.

		att_encoding_info (boolean) – Include encoding information in attachment stubs if include_docs is true and the particular attachment is compressed. Ignored if include_docs isn’t true. Default is false.

		key (json) – Return only documents that match the specified key.

		keys (json-array) – Return only documents where the key matches one of the keys specified in the array.

		sorted (boolean) – Sort returned rows (see Sorting Returned Rows). Setting this to false offers a performance boost. The total_rows and offset fields are not available when this is set to false. Default is true.

		stable (boolean) – Whether or not the view results should be returned from a stable set of shards. Default is false.

		stale (string) – Allow the results from a stale view to be used. Supported values: ok and update_after. ok is equivalent to stable=true&update=false. update_after is equivalent to stable=true&update=lazy. The default behavior is equivalent to stable=false&update=true. Note that this parameter is deprecated. Use stable and update instead. See Views Generation for more details.

		startkey_docid (string) – Return records starting with the specified document ID. Ignored if startkey is not set.

		start_key_doc_id (string) – Alias for startkey_docid param
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		got, err := newOpts(tt.options).viewOptions(tt.view)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("unexpected error: %v", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		if d := cmp.Diff(tt.want, got, cmp.AllowUnexported(viewOptions{})); d != "" {
			t.Errorf("Unexpected result:\n%s", d)
		}
	})
}
