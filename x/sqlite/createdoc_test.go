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
	"net/http"
	"regexp"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestDBCreateDoc(t *testing.T) {
	t.Parallel()
	type test struct {
		db   *testDB
		doc  interface{}
		opts driver.Options

		wantDocID  string
		wantRev    string
		wantErr    string
		wantStatus int
	}

	tests := testy.NewTable()
	tests.Add("create doc with specified doc id", test{
		doc:       map[string]string{"_id": "foo"},
		wantDocID: "foo",
		wantRev:   "1-" + (&docData{ID: "foo", Doc: []byte(`{}`)}).RevID(),
	})
	tests.Add("create doc with different doc id", test{
		doc:       map[string]string{"_id": "bar"},
		wantDocID: "bar",
		wantRev:   "1-.*",
	})
	tests.Add("invalid document should return 400", test{
		doc:        make(chan int),
		wantErr:    "json: unsupported type: chan int",
		wantStatus: http.StatusBadRequest,
	})

	/*
		TODO:
		- create doc with specific doc id, but the id already exists -- should conflict
		- create doc with no id, should generate one
		- different UUID configuration options????
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := tt.db
		if db == nil {
			db = newDB(t)
		}
		opts := tt.opts
		if opts == nil {
			opts = mock.NilOption
		}
		docID, rev, err := db.CreateDoc(context.Background(), tt.doc, opts)
		if d := errors.StatusErrorDiff(tt.wantErr, tt.wantStatus, err); d != "" {
			t.Errorf("Unexpected error: %s", d)
		}
		if err != nil {
			return
		}
		if docID != tt.wantDocID {
			t.Errorf("Unexpected doc ID. Expected %s, got %s", tt.wantDocID, docID)
		}
		if !regexp.MustCompile(tt.wantRev).MatchString(rev) {
			t.Errorf("Unexpected rev. Expected %s, got %s", tt.wantRev, rev)
		}
	})
}
