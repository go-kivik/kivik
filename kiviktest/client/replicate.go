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
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("Replicate", replicate)
}

func replicate(ctx *kt.Context) {
	defer lockReplication(ctx)()
	ctx.RunRW(func(ctx *kt.Context) {
		ctx.RunAdmin(func(ctx *kt.Context) {
			testReplication(ctx, ctx.Admin)
		})
		ctx.RunNoAuth(func(ctx *kt.Context) {
			testReplication(ctx, ctx.NoAuth)
		})
	})
}

func callReplicate(ctx *kt.Context, client *kivik.Client, target, source, repID string, opts kivik.Options) (*kivik.Replication, error) {
	opts = replicationOptions(ctx, client, target, source, repID, opts)
	var rep *kivik.Replication
	err := kt.Retry(func() error {
		var err error
		rep, err = client.Replicate(context.Background(), target, source, opts)
		return err
	})
	return rep, err
}

func testReplication(ctx *kt.Context, client *kivik.Client) {
	prefix := ctx.String("prefix")
	switch prefix {
	case "":
		prefix = strings.TrimSuffix(client.DSN(), "/") + "/"
	case "none":
		prefix = ""
	}
	targetDB, sourceDB := ctx.TestDB(), ctx.TestDB()
	defer func() {
		ctx.DestroyDB(targetDB)
		ctx.DestroyDB(sourceDB)
	}()
	dbtarget := prefix + targetDB
	dbsource := prefix + sourceDB

	db := ctx.Admin.DB(sourceDB)
	if err := db.Err(); err != nil {
		ctx.Fatalf("Failed to open db: %s", err)
	}

	// Create 10 docs for testing sync
	for i := 0; i < 10; i++ {
		id := ctx.TestDBName()
		doc := struct {
			ID string `json:"id"`
		}{
			ID: id,
		}
		if _, err := db.Put(context.Background(), doc.ID, doc); err != nil {
			ctx.Fatalf("Failed to create doc: %s", err)
		}
	}

	ctx.Run("group", func(ctx *kt.Context) {
		ctx.Run("ValidReplication", func(ctx *kt.Context) {
			ctx.Parallel()
			tries := 3
			success := false
			for i := 0; i < tries; i++ {
				success = doReplicationTest(ctx, client, dbtarget, dbsource)
				if success {
					break
				}
			}
			if !success {
				ctx.Errorf("Replication failied after %d tries", tries)
			}
		})
		ctx.Run("MissingSource", func(ctx *kt.Context) {
			ctx.Parallel()
			doReplicationTest(ctx, client, dbtarget, ctx.MustString("NotFoundDB"))
		})
		ctx.Run("MissingTarget", func(ctx *kt.Context) {
			ctx.Parallel()
			doReplicationTest(ctx, client, ctx.MustString("NotFoundDB"), dbsource)
		})
		ctx.Run("Cancel", func(ctx *kt.Context) {
			ctx.Parallel()
			replID := ctx.TestDBName()
			rep, err := callReplicate(ctx, client, dbtarget, "http://foo:foo@192.168.2.254/foo", replID, kivik.Options{"continuous": true})
			if !ctx.IsExpectedSuccess(err) {
				return
			}
			defer rep.Delete(context.Background()) // nolint: errcheck
			timeout := time.Duration(ctx.MustInt("timeoutSeconds")) * time.Second
			cx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			ctx.CheckError(rep.Delete(context.Background()))
			for rep.IsActive() {
				if rep.State() == kivik.ReplicationStarted {
					return
				}
				select {
				case <-cx.Done():
					break
				default:
				}
				if err := rep.Update(cx); err != nil {
					if kivik.HTTPStatus(err) == http.StatusNotFound {
						// NotFound expected after the replication is cancelled
						break
					}
					ctx.Fatalf("Failed to read update: %s", err)
					break
				}
			}
			if err := cx.Err(); err != nil {
				ctx.Fatalf("context was cancelled: %s", err)
			}
			if err := rep.Err(); err != nil {
				ctx.Fatalf("Replication cancellation failed: %s", err)
			}
		})
	})
}

func doReplicationTest(ctx *kt.Context, client *kivik.Client, dbtarget, dbsource string) (success bool) {
	success = true
	replID := ctx.TestDBName()
	rep, err := callReplicate(ctx, client, dbtarget, dbsource, replID, nil)
	if !ctx.IsExpectedSuccess(err) {
		return success
	}
	defer rep.Delete(context.Background()) // nolint: errcheck
	timeout := time.Duration(ctx.MustInt("timeoutSeconds")) * time.Second
	cx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	var updateErr error
	for rep.IsActive() {
		select {
		case <-cx.Done():
			ctx.Fatalf("context cancelled waiting for replication: %s", cx.Err())
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
		time.Sleep(100 * time.Millisecond)
	}
	if updateErr != nil {
		ctx.Fatalf("Replication update failed: %s", updateErr)
	}
	ctx.Run("Results", func(ctx *kt.Context) {
		err := rep.Err()
		if kivik.HTTPStatus(err) == http.StatusRequestTimeout {
			success = false // Allow retrying
			return
		}
		if !ctx.IsExpectedSuccess(err) {
			return
		}
		switch ctx.String("mode") {
		case "pouchdb":
			if rep.ReplicationID() != "" {
				ctx.Errorf("Did not expect replication ID")
			}
		default:
			if rep.State() != "completed" && rep.State() != "failed" && // 2.1.x
				rep.ReplicationID() == "" {
				ctx.Errorf("Expected a replication ID")
			}
		}
		checkReplicationURL(ctx.T, "source", dbsource, rep.Source)
		checkReplicationURL(ctx.T, "target", dbtarget, rep.Target)
		if rep.State() != kivik.ReplicationComplete {
			ctx.Errorf("Replication failed to complete. Final state: %s\n", rep.State())
		}
		if (rep.Progress() - float64(100)) > 0.0001 {
			ctx.Errorf("Expected 100%% completion, got %%%02.2f", rep.Progress())
		}
	})
	return success
}

func checkReplicationURL(t *testing.T, name, want, got string) {
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
