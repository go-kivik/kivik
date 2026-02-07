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
	"sync"
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.RegisterV2("GetReplications", getReplications)
}

// masterMU protects the map
var masterMU sync.Mutex

// We can only run one set of replication tests at a time
var replicationMUs = make(map[*kivik.Client]*sync.Mutex)

func lockReplication(c *kt.ContextCore) func() {
	masterMU.Lock()
	if _, ok := replicationMUs[c.Admin]; !ok {
		replicationMUs[c.Admin] = &sync.Mutex{}
	}
	mu := replicationMUs[c.Admin]
	masterMU.Unlock()
	mu.Lock()
	return mu.Unlock
}

func getReplications(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	defer lockReplication(c)()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		t.Parallel()
		testGetReplications(t, c, c.Admin, []struct{}{})
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		t.Parallel()
		testGetReplications(t, c, c.NoAuth, []struct{}{})
	})
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
		})
	})
}

func testGetReplications(t *testing.T, c *kt.ContextCore, client *kivik.Client, expected any) { //nolint:thelper
	reps, err := client.GetReplications(context.Background())
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	if d := testy.DiffAsJSON(expected, reps); d != nil {
		t.Errorf("GetReplications results differ:\n%s\n", d)
	}
}
