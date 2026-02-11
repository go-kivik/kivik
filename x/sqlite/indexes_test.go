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

package sqlite

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

func Test_mangoIndexName(t *testing.T) {
	t.Parallel()
	type test struct {
		dbName    string
		ddoc      string
		indexName string
		want      string
	}

	tests := testy.NewTable()
	tests.Add("standard inputs", test{
		dbName:    "testdb",
		ddoc:      "mydesigndoc",
		indexName: "myindex",
		want:      `"idx_kivik$testdb$mango_58003591"`,
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		got := mangoIndexName(tt.dbName, tt.ddoc, tt.indexName)
		if d := cmp.Diff(tt.want, got); d != "" {
			t.Errorf("Unexpected result:\n%s", d)
		}
	})
}
