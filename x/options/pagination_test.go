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

package options

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestPaginationOptions_BuildLimit(t *testing.T) {
	type test struct {
		options driver.Options
		want    string
	}

	tests := testy.NewTable()
	tests.Add("no limit or skip", test{
		options: mock.NilOption,
		want:    "",
	})
	tests.Add("limit only", test{
		options: kivik.Param("limit", 10),
		want:    "LIMIT 10",
	})
	tests.Add("skip only", test{
		options: kivik.Param("skip", 5),
		want:    "LIMIT -1 OFFSET 5",
	})
	tests.Add("limit and skip", test{
		options: kivik.Params(map[string]interface{}{
			"limit": 10,
			"skip":  5,
		}),
		want: "LIMIT 10 OFFSET 5",
	})
	tests.Add("limit=0", test{
		options: kivik.Param("limit", 0),
		want:    "LIMIT 0",
	})

	tests.Run(t, func(t *testing.T, tt test) {
		pagination, err := New(tt.options).PaginationOptions(true)
		if err != nil {
			t.Fatal(err)
		}
		got := pagination.BuildLimit()
		if d := cmp.Diff(tt.want, got); d != "" {
			t.Errorf("Unexpected result:\n%s", d)
		}
	})
}
