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
	"context"
	"crypto/md5"
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

type DB interface {
	driver.DB
	driver.Purger
	driver.RevsDiffer
	driver.OpenRever
	driver.AttachmentMetaGetter
}

type testDB struct {
	t *testing.T
	DB
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
	client, err := d.NewClient(dsn, nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := client.CreateDB(context.Background(), "test", nil); err != nil {
		t.Fatal(err)
	}
	db, err := client.DB("test", nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})
	return &testDB{
		DB: db.(DB),
		t:  t,
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
