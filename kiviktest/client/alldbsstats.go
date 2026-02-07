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
	kt.RegisterV2("AllDBsStats", allDBsStats)
}

func allDBsStats(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		testAllDBsStats(t, c, c.Admin)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		testAllDBsStats(t, c, c.NoAuth)
	})
}

func testAllDBsStats(t *testing.T, c *kt.ContextCore, client *kivik.Client) { //nolint:thelper
	dbName := c.TestDB(t)
	stats, err := client.AllDBsStats(context.Background())
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	for _, stat := range stats {
		if stat.Name == dbName {
			return
		}
	}
	t.Errorf("Expected database %s to be in stats, but it was not found", dbName)
}
