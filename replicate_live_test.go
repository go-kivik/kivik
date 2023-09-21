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

package xkivik

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb" // CouchDB driver
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
	_ "github.com/go-kivik/kivik/v4/x/fsdb" // Filesystem driver
)

func TestReplicate_live(t *testing.T) { //nolint:gocyclo // allowed for subtests
	type tt struct {
		source, target *kivik.DB
		options        kivik.Option
		status         int
		err            string
		result         *ReplicationResult
	}
	tests := testy.NewTable()
	tests.Add("couch to couch", func(t *testing.T) interface{} {
		dsn := kt.DSN(t)
		client, err := kivik.New("couch", dsn)
		if err != nil {
			t.Fatal(err)
		}
		sourceName := kt.TestDBName(t)
		targetName := kt.TestDBName(t)
		ctx := context.Background()
		if err := client.CreateDB(ctx, sourceName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, sourceName)
		})
		if err := client.CreateDB(ctx, targetName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, targetName)
		})
		source := client.DB(sourceName)
		target := client.DB(targetName)
		doc := map[string]string{"foo": "bar"}
		if _, err := source.Put(ctx, "foo", doc); err != nil {
			t.Fatal(err)
		}

		return tt{
			source: source,
			target: target,
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})
	tests.Add("fs to couch", func(t *testing.T) interface{} {
		fsclient, err := kivik.New("fs", "testdata/")
		if err != nil {
			t.Fatal(err)
		}
		dsn := kt.DSN(t)
		client, err := kivik.New("couch", dsn)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		source := fsclient.DB("db1")
		targetName := kt.TestDBName(t)
		if err := client.CreateDB(ctx, targetName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, targetName)
		})
		target := client.DB(targetName)

		return tt{
			source: source,
			target: target,
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})
	tests.Add("fs to couch, no shared history", func(t *testing.T) interface{} {
		fsclient, err := kivik.New("fs", "testdata/")
		if err != nil {
			t.Fatal(err)
		}
		dsn := kt.DSN(t)
		client, err := kivik.New("couch", dsn)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		source := fsclient.DB("db1")
		targetName := kt.TestDBName(t)
		if err := client.CreateDB(ctx, targetName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, targetName)
		})
		target := client.DB(targetName)

		if _, err := Replicate(ctx, target, source); err != nil {
			t.Fatalf("setup replication failed: %s", err)
		}

		return tt{
			source: fsclient.DB("db2"),
			target: target,
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})
	tests.Add("couch to couch with sec", func(t *testing.T) interface{} {
		dsn := kt.DSN(t)
		client, err := kivik.New("couch", dsn)
		if err != nil {
			t.Fatal(err)
		}
		sourceName := kt.TestDBName(t)
		targetName := kt.TestDBName(t)
		ctx := context.Background()
		if err := client.CreateDB(ctx, sourceName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, sourceName)
		})
		if err := client.CreateDB(ctx, targetName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, targetName)
		})
		source := client.DB(sourceName)
		target := client.DB(targetName)
		doc := map[string]string{"foo": "bar"}
		if _, err := source.Put(ctx, "foo", doc); err != nil {
			t.Fatal(err)
		}
		err = source.SetSecurity(ctx, &kivik.Security{
			Members: kivik.Members{
				Names: []string{"bob"},
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		return tt{
			source:  source,
			target:  target,
			options: kivik.Param("copy_security", true),
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})
	tests.Add("fs to couch, bad put", func(t *testing.T) interface{} {
		fsclient, err := kivik.New("fs", "testdata/")
		if err != nil {
			t.Fatal(err)
		}
		dsn := kt.DSN(t)
		client, err := kivik.New("couch", dsn)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		targetName := kt.TestDBName(t)
		if err := client.CreateDB(ctx, targetName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, targetName)
		})
		target := client.DB(targetName)

		return tt{
			source: fsclient.DB("db3"),
			target: target,
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
			status: http.StatusBadRequest,
			err:    "store doc note--XkWjFv13acvjJTt-CGJJ8hXlWE: Bad Request: Bad special document member: _invalid",
		}
	})
	tests.Add("fs to couch with attachment", func(t *testing.T) interface{} {
		fsclient, err := kivik.New("fs", "testdata/")
		if err != nil {
			t.Fatal(err)
		}
		dsn := kt.DSN(t)
		client, err := kivik.New("couch", dsn)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		source := fsclient.DB("db4")
		targetName := kt.TestDBName(t)
		if err := client.CreateDB(ctx, targetName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, targetName)
		})
		target := client.DB(targetName)

		return tt{
			source: source,
			target: target,
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})
	tests.Add("couch to fs", func(t *testing.T) interface{} {
		tempDir, err := os.MkdirTemp("", "kivik.test.")
		if err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() error {
			return os.RemoveAll(tempDir)
		})
		tClient, err := kivik.New("fs", tempDir)
		if err != nil {
			t.Fatal(err)
		}

		dsn := kt.DSN(t)
		client, err := kivik.New("couch", dsn)
		if err != nil {
			t.Fatal(err)
		}
		dbName := kt.TestDBName(t)
		ctx := context.Background()
		if err := client.CreateDB(ctx, dbName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, dbName)
		})
		if err := tClient.CreateDB(ctx, dbName); err != nil {
			t.Fatal(err)
		}
		source := client.DB(dbName)
		target := tClient.DB(dbName)
		doc := map[string]interface{}{
			"foo": "bar",
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "application/octet-stream",
					"data":         []byte("Test content"),
				},
			},
		}
		if _, err := source.Put(ctx, "foo", doc); err != nil {
			t.Fatal(err)
		}

		return tt{
			source: source,
			target: target,
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})
	tests.Add("fs to couch with deleted document", func(t *testing.T) interface{} {
		fsclient, err := kivik.New("fs", "testdata/")
		if err != nil {
			t.Fatal(err)
		}
		dsn := kt.DSN(t)
		client, err := kivik.New("couch", dsn)
		if err != nil {
			t.Fatal(err)
		}
		ctx := context.Background()
		source := fsclient.DB("dbdelete")
		targetName := kt.TestDBName(t)
		if err := client.CreateDB(ctx, targetName); err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = client.DestroyDB(ctx, targetName)
		})
		target := client.DB(targetName)
		if _, err := target.Put(ctx, "foo", map[string]string{"still": "here"}); err != nil {
			t.Fatal(err)
		}

		return tt{
			source: source,
			target: target,
			result: &ReplicationResult{
				DocsRead:       1,
				DocsWritten:    1,
				MissingChecked: 1,
				MissingFound:   1,
			},
		}
	})
	tests.Run(t, func(t *testing.T, tt tt) {
		ctx := context.TODO()
		result, err := Replicate(ctx, tt.target, tt.source, tt.options)
		testy.StatusError(t, tt.err, tt.status, err)

		verifyDoc(ctx, t, tt.target, tt.source, "foo")
		verifySec(ctx, t, tt.target)
		result.StartTime = time.Time{}
		result.EndTime = time.Time{}
		if d := testy.DiffAsJSON(tt.result, result); d != nil {
			t.Error(d)
		}
	})
}

func verifyDoc(ctx context.Context, t *testing.T, target, source *kivik.DB, docID string) {
	t.Helper()
	var targetDoc, sourceDoc interface{}
	notFound := false
	if err := source.Get(ctx, docID).ScanDoc(&sourceDoc); err != nil {
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			notFound = true
		} else {
			t.Fatalf("get %s from source failed: %s", docID, err)
		}
	}
	if err := target.Get(ctx, docID).ScanDoc(&targetDoc); err != nil {
		if notFound && kivik.HTTPStatus(err) == http.StatusNotFound {
			return
		}
		t.Fatalf("get %s from target failed: %s", docID, err)
	}
	if d := testy.DiffAsJSON(sourceDoc, targetDoc); d != nil {
		t.Error(d)
	}
}

func verifySec(ctx context.Context, t *testing.T, target *kivik.DB) {
	sec, err := target.Security(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if d := testy.DiffAsJSON(&testy.File{Path: "testdata/" + testy.Stub(t) + ".security"}, sec); d != nil {
		t.Errorf("Security object:\n%s", d)
	}
}
