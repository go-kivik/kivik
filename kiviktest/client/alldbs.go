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

// Package client provides integration tests for the kivik client.
package client

import (
	"context"
	"sort"
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("AllDBs", allDBs)
}

func allDBs(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		testAllDBs(t, c, c.Admin, c.StringSlice(t, "expected"))
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		testAllDBs(t, c, c.NoAuth, c.StringSlice(t, "expected"))
	})
	if c.RW && c.Admin != nil {
		c.Run(t, "RW", func(t *testing.T) {
			t.Helper()
			testAllDBsRW(t, c)
		})
	}
}

func testAllDBsRW(t *testing.T, c *kt.Context) { //nolint:thelper
	dbName := c.TestDB(t)
	expected := append(c.StringSlice(t, "expected"), dbName)
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		testAllDBs(t, c, c.Admin, expected)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		testAllDBs(t, c, c.NoAuth, expected)
	})
}

func testAllDBs(t *testing.T, c *kt.Context, client *kivik.Client, expected []string) { //nolint:thelper
	t.Parallel()
	allDBs, err := client.AllDBs(context.Background())
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	sort.Strings(expected)
	sort.Strings(allDBs)
	if d := testy.DiffTextSlices(expected, allDBs); d != nil {
		t.Errorf("AllDBs() returned unexpected list:\n%s\n", d)
	}
}
