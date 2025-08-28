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

	"github.com/google/go-cmp/cmp"

	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestClientAllDBs(t *testing.T) {
	d := drv{}
	dClient, err := d.NewClient(":memory:", mock.NilOption)
	if err != nil {
		t.Fatal(err)
	}

	c := dClient.(*client)
	for _, table := range []string{"foo", "bar", "foo_attachments", "bar_attachments", "foo_attachments_bridge", "foo_design", "bar_design", "foo_revs", "bar_revs"} {
		if _, err := c.db.Exec("CREATE TABLE " + table + " (id INTEGER)"); err != nil {
			t.Fatal(err)
		}
	}

	dbs, err := dClient.AllDBs(context.Background(), mock.NilOption)
	if err != nil {
		t.Fatal("err should be nil")
	}
	wantDBs := []string{"foo", "bar"}
	if d := cmp.Diff(wantDBs, dbs); d != "" {
		t.Fatal(d)
	}
}
