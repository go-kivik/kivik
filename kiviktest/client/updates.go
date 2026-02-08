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
	"fmt"
	"testing"
	"time"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("DBUpdates", updates)
}

const maxWait = 5 * time.Second

func updates(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			testUpdates(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			testUpdates(t, c, c.NoAuth)
		})
	})
}

func testUpdates(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	t.Parallel()
	updates := client.DBUpdates(context.TODO())
	if !c.IsExpectedSuccess(t, updates.Err()) {
		return
	}
	// It seems that DBUpdates doesn't always start responding immediately,
	// so introduce a small delay to ensure we're reporting updates before we
	// actually do the updates.
	const delay = 10 * time.Millisecond
	time.Sleep(delay)
	dbname := kt.TestDBName(t)
	eventErrors := make(chan error)
	go func() {
		for updates.Next() {
			if updates.DBName() == dbname {
				if updates.Type() == "created" {
					break
				}
				eventErrors <- fmt.Errorf("unexpected event type '%s'", updates.Type())
			}
		}
		eventErrors <- updates.Err()
		close(eventErrors)
	}()
	t.Cleanup(func() { c.DestroyDB(t, dbname) })
	if err := c.Admin.CreateDB(context.Background(), dbname, c.Options(t, "db")); err != nil {
		t.Fatalf("Failed to create db: %s", err)
	}
	timer := time.NewTimer(maxWait)
	select {
	case err := <-eventErrors:
		if err != nil {
			t.Fatalf("Error reading event: %s", err)
		}
	case <-timer.C:
		t.Fatalf("Failed to read expected event in %s", maxWait)
	}
	if err := updates.Close(); err != nil {
		t.Errorf("Updates close failed: %s", err)
	}
}
