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
		},
	})
	tests.Add("confclits: invalid string", test{
		options:    kivik.Param("conflicts", "oink"),
		wantErr:    "invalid value for `conflicts`: oink",
		wantStatus: http.StatusBadRequest,
	})
	tests.Add("conflicts: string", test{
		options: kivik.Param("conflicts", "true"),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			conflicts:    true,
		},
	})
	tests.Add("conflicts: bool", test{
		options: kivik.Param("conflicts", true),
		want: &viewOptions{
			limit:        -1,
			inclusiveEnd: true,
			conflicts:    true,
		},
	})
	tests.Add("conflicts: int", test{
		options:    kivik.Param("conflicts", 3),
		wantErr:    "invalid value for `conflicts`: 3",
		wantStatus: http.StatusBadRequest,
	})

	/*
		descending (boolean) – Return the documents in descending order by key. Default is false.

		endkey (json) – Stop returning records when the specified key is reached.

		end_key (json) – Alias for endkey param

		endkey_docid (string) – Stop returning records when the specified document ID is reached. Ignored if endkey is not set.

		end_key_doc_id (string) – Alias for endkey_docid.

		group (boolean) – Group the results using the reduce function to a group or single row. Implies reduce is true and the maximum group_level. Default is false.

		group_level (number) – Specify the group level to be used. Implies group is true.

		include_docs (boolean) – Include the associated document with each row. Default is false.

		attachments (boolean) – Include the Base64-encoded content of attachments in the documents that are included if include_docs is true. Ignored if include_docs isn’t true. Default is false.

		att_encoding_info (boolean) – Include encoding information in attachment stubs if include_docs is true and the particular attachment is compressed. Ignored if include_docs isn’t true. Default is false.

		inclusive_end (boolean) – Specifies whether the specified end key should be included in the result. Default is true.

		key (json) – Return only documents that match the specified key.

		keys (json-array) – Return only documents where the key matches one of the keys specified in the array.

		limit (number) – Limit the number of the returned documents to the specified number.

		reduce (boolean) – Use the reduction function. Default is true when a reduce function is defined.

		skip (number) – Skip this number of records before starting to return the results. Default is 0.

		sorted (boolean) – Sort returned rows (see Sorting Returned Rows). Setting this to false offers a performance boost. The total_rows and offset fields are not available when this is set to false. Default is true.

		stable (boolean) – Whether or not the view results should be returned from a stable set of shards. Default is false.

		stale (string) – Allow the results from a stale view to be used. Supported values: ok and update_after. ok is equivalent to stable=true&update=false. update_after is equivalent to stable=true&update=lazy. The default behavior is equivalent to stable=false&update=true. Note that this parameter is deprecated. Use stable and update instead. See Views Generation for more details.

		startkey (json) – Return records starting with the specified key.

		start_key (json) – Alias for startkey.

		startkey_docid (string) – Return records starting with the specified document ID. Ignored if startkey is not set.

		start_key_doc_id (string) – Alias for startkey_docid param

		update (string) – Whether or not the view in question should be updated prior to responding to the user. Supported values: true, false, lazy. Default is true.

		update_seq (boolean) – Whether to include in the response an update_seq value indicating the sequence id of the database the view reflects. Default is false.
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
