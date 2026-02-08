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
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("DeleteAttachment", delAttachment)
}

func delAttachment(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		dbname := c.TestDB(t)
		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testDeleteAttachments(t, c, c.Admin, dbname, "foo.txt")
			testDeleteAttachments(t, c, c.Admin, dbname, "NotFound")
			testDeleteAttachmentsDDoc(t, c, c.Admin, dbname, "foo.txt")
			testDeleteAttachmentNoDoc(t, c, c.Admin, dbname)
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testDeleteAttachments(t, c, c.NoAuth, dbname, "foo.txt")
			testDeleteAttachments(t, c, c.NoAuth, dbname, "NotFound")
			testDeleteAttachmentsDDoc(t, c, c.NoAuth, dbname, "foo.txt")
			testDeleteAttachmentNoDoc(t, c, c.NoAuth, dbname)
		})
	})
}

func testDeleteAttachmentNoDoc(t *testing.T, c *kt.Context, client *kivik.Client, dbname string) { //nolint:thelper
	db := c.DB(t, client, dbname)
	c.Run(t, "NoDoc", func(t *testing.T) {
		t.Parallel()
		_, err := db.DeleteAttachment(context.Background(), "nonexistantdoc", "2-4259cd84694a6345d6c534ed65f1b30b", "foo.txt")
		c.CheckError(t, err)
	})
}

func testDeleteAttachments(t *testing.T, c *kt.Context, client *kivik.Client, dbname, filename string) { //nolint:thelper
	c.Run(t, filename, func(t *testing.T) {
		doDeleteAttachmentTest(t, c, client, dbname, kt.TestDBName(t), filename)
	})
}

func testDeleteAttachmentsDDoc(t *testing.T, c *kt.Context, client *kivik.Client, dbname, filename string) { //nolint:thelper
	c.Run(t, "DesignDoc/"+filename, func(t *testing.T) {
		doDeleteAttachmentTest(t, c, client, dbname, "_design/"+kt.TestDBName(t), filename)
	})
}

func doDeleteAttachmentTest(t *testing.T, c *kt.Context, client *kivik.Client, dbname, docID, filename string) { //nolint:thelper
	db := c.DB(t, client, dbname)
	t.Parallel()
	adb := c.AdminDB(t, dbname)
	doc := map[string]any{
		"_id": docID,
		"_attachments": map[string]any{
			"foo.txt": map[string]any{
				"content_type": "text/plain",
				"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGVkIHRleHQ=",
			},
		},
	}
	rev, err := adb.Put(context.Background(), docID, doc)
	if err != nil {
		t.Fatalf("Failed to create doc: %s", err)
	}
	rev, err = db.DeleteAttachment(context.Background(), docID, rev, filename)
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	var i any
	if err = db.Get(context.Background(), docID, kivik.Rev(rev)).ScanDoc(&i); err != nil {
		t.Fatalf("Failed to get deleted doc: %s", err)
	}
}
