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
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("Replicate", replicate)
}

func replicate(t *testing.T, c *kt.Context) {
	t.Helper()
	defer lockReplication(c)()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			testReplication(t, c, c.Admin)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			testReplication(t, c, c.NoAuth)
		})
	})
}

func callReplicate(t *testing.T, c *kt.Context, client *kivik.Client, target, source, repID string, options kivik.Option) (*kivik.Replication, error) { //nolint:thelper
	options = replicationOptions(t, c, target, source, repID, options)
	var rep *kivik.Replication
	err := kt.Retry(func() error {
		var err error
		rep, err = client.Replicate(context.Background(), target, source, options)
		return err
	})
	return rep, err
}

func testReplication(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	prefix := c.String(t, "prefix")
	switch prefix {
	case "":
		prefix = strings.TrimSuffix(client.DSN(), "/") + "/"
	case "none":
		prefix = ""
	}
	targetDB, sourceDB := c.TestDB(t), c.TestDB(t)
	dbtarget := prefix + targetDB
	dbsource := prefix + sourceDB

	db := c.Admin.DB(sourceDB)
	if err := db.Err(); err != nil {
		t.Fatalf("Failed to open db: %s", err)
	}

	// Create 10 docs for testing sync
	for i := 0; i < 10; i++ {
		id := kt.TestDBName(t)
		doc := struct {
			ID string `json:"id"`
		}{
			ID: id,
		}
		if _, err := db.Put(context.Background(), doc.ID, doc); err != nil {
			t.Fatalf("Failed to create doc: %s", err)
		}
	}

	c.Run(t, "ValidReplication", func(t *testing.T) {
		t.Parallel()
		tries := 3
		success := false
		for i := 0; i < tries; i++ {
			success = doReplicationTest(t, c, client, dbtarget, dbsource)
			if success {
				break
			}
		}
		if !success {
			t.Errorf("Replication failed after %d tries", tries)
		}
	})
	c.Run(t, "MissingSource", func(t *testing.T) {
		t.Parallel()
		doReplicationTest(t, c, client, dbtarget, c.MustString(t, "NotFoundDB"))
	})
	c.Run(t, "MissingTarget", func(t *testing.T) {
		t.Parallel()
		doReplicationTest(t, c, client, c.MustString(t, "NotFoundDB"), dbsource)
	})
	c.Run(t, "Cancel", func(t *testing.T) {
		t.Parallel()
		replID := kt.TestDBName(t)
		rep, err := callReplicate(t, c, client, dbtarget, "http://foo:foo@192.168.2.254/foo", replID, kivik.Param("continuous", true))
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		t.Cleanup(func() { _ = rep.Delete(context.Background()) })
		timeout := time.Duration(c.MustInt(t, "timeoutSeconds")) * time.Second
		cx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()
		c.CheckError(t, rep.Delete(context.Background()))
	loop:
		for rep.IsActive() {
			if rep.State() == kivik.ReplicationStarted {
				return
			}
			select {
			case <-cx.Done():
				break loop
			default:
			}
			if err := rep.Update(cx); err != nil {
				if kivik.HTTPStatus(err) == http.StatusNotFound {
					// NotFound expected after the replication is cancelled
					break
				}
				t.Fatalf("Failed to read update: %s", err)
			}
		}
		if err := cx.Err(); err != nil {
			t.Fatalf("context was cancelled: %s", err)
		}
		if err := rep.Err(); err != nil {
			t.Fatalf("Replication cancellation failed: %s", err)
		}
	})
}

func doReplicationTest(t *testing.T, c *kt.Context, client *kivik.Client, dbtarget, dbsource string) (success bool) { //nolint:thelper
	success = true
	replID := kt.TestDBName(t)
	rep, err := callReplicate(t, c, client, dbtarget, dbsource, replID, nil)
	if !c.IsExpectedSuccess(t, err) {
		return success
	}
	t.Cleanup(func() { _ = rep.Delete(context.Background()) })
	timeout := time.Duration(c.MustInt(t, "timeoutSeconds")) * time.Second
	cx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var updateErr error
	for rep.IsActive() {
		select {
		case <-cx.Done():
			t.Fatalf("context cancelled waiting for replication: %s", cx.Err())
			return success
		default:
		}
		if updateErr = rep.Update(cx); updateErr != nil {
			break
		}
		if rep.State() == "crashing" {
			// 2.1 treats Not Found as a temporary error (on the theory the missing
			// db could be created), so this short-circuits.
			break
		}
		const delay = 100 * time.Millisecond
		time.Sleep(delay)
	}
	if updateErr != nil {
		t.Fatalf("Replication update failed: %s", updateErr)
	}
	c.Run(t, "Results", func(t *testing.T) {
		err := rep.Err()
		if kivik.HTTPStatus(err) == http.StatusRequestTimeout {
			success = false // Allow retrying
			return
		}
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		switch c.String(t, "mode") {
		case "pouchdb":
			if rep.ReplicationID() != "" {
				t.Errorf("Did not expect replication ID")
			}
		default:
			if rep.State() != "completed" && rep.State() != "failed" && // 2.1.x
				rep.ReplicationID() == "" {
				t.Errorf("Expected a replication ID")
			}
		}
		checkReplicationURL(t, "source", dbsource, rep.Source)
		checkReplicationURL(t, "target", dbtarget, rep.Target)
		if rep.State() != kivik.ReplicationComplete {
			t.Errorf("Replication failed to complete. Final state: %s\n", rep.State())
		}
		const (
			pct100  = float64(100)
			minProg = 0.0001
		)
		if (rep.Progress() - pct100) > minProg {
			t.Errorf("Expected 100%% completion, got %%%02.2f", rep.Progress())
		}
	})
	return success
}

func checkReplicationURL(t *testing.T, name, want, got string) {
	t.Helper()
	wantURL, err := url.Parse(want)
	if err != nil {
		t.Fatal(err)
	}
	gotURL, err := url.Parse(got)
	if err != nil {
		t.Fatal(err)
	}
	if !replicationURLsEqual(wantURL, gotURL) {
		t.Errorf("Unexpected %s URL. Want: %s, got %s", name, want, got)
	}
}

func replicationURLsEqual(want, got *url.URL) bool {
	if want.User != nil && got.User != nil {
		wantUser := want.User.Username()
		gotUser := got.User.Username()
		if wantUser != "" && gotUser != "" && wantUser != gotUser {
			return false
		}
	}
	want.User = nil
	got.User = nil
	want.Path = filepath.Join(want.Path, "")
	got.Path = filepath.Join(got.Path, "")
	return want.String() == got.String()
}
