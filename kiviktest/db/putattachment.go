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

package db

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.RegisterV2("PutAttachment", putAttachment)
}

func putAttachment(t *testing.T, c *kt.ContextCore) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		dbname := c.TestDB(t)
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testPutAttachment(t, c, c.Admin, dbname)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testPutAttachment(t, c, c.NoAuth, dbname)
		})
	})
}

func testPutAttachment(t *testing.T, c *kt.ContextCore, client *kivik.Client, dbname string) { //nolint:thelper
	db := client.DB(dbname, c.Options(t, "db"))
	if err := db.Err(); err != nil {
		t.Fatalf("Failed to open db: %s", err)
	}
	adb := c.Admin.DB(dbname, c.Options(t, "db"))
	if err := adb.Err(); err != nil {
		t.Fatalf("Failed to open admin db: %s", err)
	}
	c.Run(t, "Update", func(t *testing.T) {
		t.Parallel()
		var docID, rev string
		err := kt.Retry(func() error {
			var e error
			docID, rev, e = adb.CreateDoc(context.Background(), map[string]string{"name": "Robert"})
			return e
		})
		if err != nil {
			t.Fatalf("Failed to create doc: %s", err)
		}
		err = kt.Retry(func() error {
			att := &kivik.Attachment{
				Filename:    "test.txt",
				ContentType: "text/plain",
				Content:     stringReadCloser(),
			}
			_, err = db.PutAttachment(context.Background(), docID, att, kivik.Rev(rev))
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "Create", func(t *testing.T) {
		t.Parallel()
		docID := kt.TestDBName(t)
		err := kt.Retry(func() error {
			att := &kivik.Attachment{
				Filename:    "test.txt",
				ContentType: "text/plain",
				Content:     stringReadCloser(),
			}
			_, err := db.PutAttachment(context.Background(), docID, att)
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "Conflict", func(t *testing.T) {
		t.Parallel()
		var docID string
		err2 := kt.Retry(func() error {
			var e error
			docID, _, e = adb.CreateDoc(context.Background(), map[string]string{"name": "Robert"})
			return e
		})
		if err2 != nil {
			t.Fatalf("Failed to create doc: %s", err2)
		}
		err := kt.Retry(func() error {
			att := &kivik.Attachment{
				Filename:    "test.txt",
				ContentType: "text/plain",
				Content:     stringReadCloser(),
			}
			_, err := db.PutAttachment(context.Background(), docID, att, kivik.Rev("5-20bd3c7d7d6b81390c6679d8bae8795b"))
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "UpdateDesignDoc", func(t *testing.T) {
		t.Parallel()
		docID := "_design/" + kt.TestDBName(t)
		doc := map[string]string{
			"_id": docID,
		}
		var rev string
		err := kt.Retry(func() error {
			var err error
			rev, err = adb.Put(context.Background(), docID, doc)
			return err
		})
		if err != nil {
			t.Fatalf("Failed to create design doc: %s", err)
		}
		err = kt.Retry(func() error {
			att := &kivik.Attachment{
				Filename:    "test.txt",
				ContentType: "text/plain",
				Content:     stringReadCloser(),
			}
			_, err = db.PutAttachment(context.Background(), docID, att, kivik.Rev(rev))
			return err
		})
		c.CheckError(t, err)
	})
	c.Run(t, "CreateDesignDoc", func(t *testing.T) {
		t.Parallel()
		docID := "_design/" + kt.TestDBName(t)
		err := kt.Retry(func() error {
			att := &kivik.Attachment{
				Filename:    "test.txt",
				ContentType: "text/plain",
				Content:     stringReadCloser(),
			}
			_, err := db.PutAttachment(context.Background(), docID, att)
			return err
		})
		c.CheckError(t, err)
	})
}

func stringReadCloser() io.ReadCloser {
	return io.NopCloser(strings.NewReader("test content"))
}
