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
// +build !js

package sqlite

import (
	"bytes"
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

type DB interface {
	driver.DB
	driver.Purger
	driver.RevsDiffer
	driver.OpenRever
	driver.AttachmentMetaGetter
	driver.RevGetter
	driver.LocalDocer
	driver.DesignDocer
	driver.DocCreator
}

type testDB struct {
	t *testing.T
	DB
	logs *bytes.Buffer
}

func (tdb *testDB) underlying() *sql.DB {
	return tdb.DB.(*db).db
}

func (tdb *testDB) tPut(docID string, doc interface{}, options ...driver.Options) string {
	tdb.t.Helper()
	opt := driver.Options(mock.NilOption)
	if len(options) > 0 {
		opt = options[0]
	}
	rev, err := tdb.Put(context.Background(), docID, doc, opt)
	if err != nil {
		tdb.t.Fatalf("Failed to put doc: %s", err)
	}
	return rev
}

func (tdb *testDB) tDelete(docID string, options ...driver.Options) string { //nolint:unparam
	tdb.t.Helper()
	opt := driver.Options(mock.NilOption)
	if len(options) > 0 {
		opt = options[0]
	}
	rev, err := tdb.Delete(context.Background(), docID, opt)
	if err != nil {
		tdb.t.Fatalf("Failed to delete doc: %s", err)
	}
	return rev
}

// newDB creates a new driver.DB instance backed by an in-memory SQLite database,
// and registers a cleanup function to close the database when the test is done.
func newDB(t *testing.T) *testDB {
	var dsn string
	if os.Getenv("KEEP_TEST_DB") != "" {
		file, err := os.CreateTemp("", "kivik-sqlite-test-*.db")
		if err != nil {
			t.Fatal(err)
		}
		dsn = file.Name()
		if err := file.Close(); err != nil {
			t.Fatal(err)
		}
		t.Logf("Test database: %s", dsn)
	} else {
		// calculate md5sum of test name
		md5sum := md5.New()
		_, _ = md5sum.Write([]byte(t.Name()))
		dsn = fmt.Sprintf("file:%x?mode=memory&cache=shared", md5sum.Sum(nil))
	}
	d := drv{}
	buf := &bytes.Buffer{}
	logger := log.New(buf, "", 0)
	client, err := d.NewClient(dsn, OptionLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CreateDB(context.Background(), "test", nil); err != nil {
		t.Fatal(err)
	}
	db, err := client.DB("test", mock.NilOption)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return &testDB{
		DB:   db.(DB),
		t:    t,
		logs: buf,
	}
}

func (tdb *testDB) checkLogs(want []string) {
	tdb.t.Helper()
	if want == nil {
		return
	}
	got := strings.Split(tdb.logs.String(), "\n")
	if len(got) > 0 && got[len(got)-1] == "" {
		got = got[:len(got)-1]
	}
	for i := 0; i < len(got) && i < len(want); i++ {
		re := regexp.MustCompile(want[i])
		if !re.MatchString(got[i]) {
			tdb.t.Errorf("Unexpected log line: %d. Want /%s/,\n\t Got %s", i+1, want[i], got[i])
		}
	}
	if len(got) > len(want) {
		tdb.t.Errorf("Got %d more logs than expected: %s", len(got)-len(want), strings.Join(got[len(want):], "\n"))
	}
	if len(got) < len(want) {
		tdb.t.Errorf("Got %d fewer logs than expected", len(want)-len(got))
	}
}

type testAttachments map[string]interface{}

// newAttachments returns a new testAttachments map. Use [add] to add one or
// more attachments.
func newAttachments() testAttachments {
	return make(testAttachments)
}

func (a testAttachments) add(filename, content string) testAttachments {
	a[filename] = map[string]interface{}{
		"content_type": "text/plain",
		"data":         []byte(content),
	}
	return a
}

func (a testAttachments) addStub(filename string) testAttachments {
	a[filename] = map[string]interface{}{
		"stub": true,
	}
	return a
}

type reduced struct {
	Seq      int
	Depth    int
	FirstKey string
	FirstPK  int
	LastKey  string
	LastPK   int
	Value    string
}

func checkReduced(t *testing.T, db *sql.DB, want []reduced) {
	t.Helper()
	var table string
	if err := db.QueryRow(`
		SELECT name
		FROM sqlite_master
		WHERE type = 'table'
			AND name LIKE '%_%_reduce_%'
	`).Scan(&table); err != nil {
		t.Fatalf("Failed to find reduced table: %s", err)
	}
	rows, err := db.Query(fmt.Sprintf(`
		SELECT *
		FROM %q
		ORDER BY first_key, first_pk, last_key, last_pk, depth`, table))
	if err != nil {
		t.Fatalf("Failed to query reduced table: %s", err)
	}
	defer rows.Close()
	got := make([]reduced, 0, len(want))
	for rows.Next() {
		var r reduced
		if err := rows.Scan(&r.Seq, &r.Depth, &r.FirstKey, &r.FirstPK, &r.LastKey, &r.LastPK, &r.Value); err != nil {
			t.Fatalf("Failed to scan reduced row: %s", err)
		}
		got = append(got, r)
	}
	if err := rows.Err(); err != nil {
		t.Fatalf("Failed to iterate reduced rows: %s", err)
	}
	if d := cmp.Diff(want, got); d != "" {
		t.Errorf("Unexpected reduced rows (-want, +got): %s", d)
	}
}
