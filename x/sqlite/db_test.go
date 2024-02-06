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
	"context"
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestDBPut(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		docID      string
		doc        interface{}
		wantRev    string
		wantStatus int
		wantErr    string
	}{
		{
			name:  "create new document",
			docID: "foo",
			doc: map[string]string{
				"foo": "bar",
			},
			wantRev: "1-6fe51f74859f3579abaccc426dd5104f",
		},
		/*
			new document, with rev: create doc with provided rev (verify)
			existing doc, no rev: conflict
			existing document, with matching rev: create new rev
			existing document, non-matching rev: conflict
			existing document, new_edits: Accept any rev as-is
			new_edits, missing rev: ??
		*/
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := drv{}
			client, err := d.NewClient(":memory:", nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := client.CreateDB(context.Background(), "test", nil); err != nil {
				t.Fatal(err)
			}
			db, err := client.DB("test", nil)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				_ = db.Close()
			})
			rev, err := db.Put(context.Background(), tt.docID, tt.doc, nil)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if err != nil {
				return
			}
			if rev != tt.wantRev {
				t.Errorf("Unexpected rev: %s, want %s", rev, tt.wantRev)
			}
		})
	}
}
