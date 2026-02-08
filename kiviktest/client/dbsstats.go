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

package client

import (
	"context"
	"testing"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("DBsStats", dbsStats)
}

func dbsStats(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		testDBsStats(t, c, c.Admin)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		testDBsStats(t, c, c.NoAuth)
	})
}

func testDBsStats(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	stats, err := client.DBsStats(context.Background(), []string{"_users", "notfound"})
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	const wantResults = 2
	if len(stats) != wantResults {
		t.Errorf("Expected %d database stats, got %d", wantResults, len(stats))
	}
}
