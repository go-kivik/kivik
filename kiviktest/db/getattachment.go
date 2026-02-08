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
	kt.Register("GetAttachment", getAttachment)
}

func getAttachment(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunRW(t, func(t *testing.T) {
		t.Helper()
		dbname := c.TestDB(t)
		adb := c.AdminDB(t, dbname)

		doc := map[string]any{
			"_id": "foo",
			"_attachments": map[string]any{
				"foo.txt": map[string]any{
					"content_type": "text/plain",
					"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGVkIHRleHQ=",
				},
			},
		}
		if _, err := adb.Put(context.Background(), "foo", doc); err != nil {
			t.Fatalf("Failed to create doc: %s", err)
		}

		ddoc := map[string]any{
			"_id": "_design/foo",
			"_attachments": map[string]any{
				"foo.txt": map[string]any{
					"content_type": "text/plain",
					"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGVkIHRleHQ=",
				},
			},
		}
		if _, err := adb.Put(context.Background(), "_design/foo", ddoc); err != nil {
			t.Fatalf("Failed to create design doc: %s", err)
		}

		c.RunAdmin(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testGetAttachments(t, c, c.Admin, dbname, "foo", "foo.txt")
			testGetAttachments(t, c, c.Admin, dbname, "foo", "NotFound")
			testGetAttachments(t, c, c.Admin, dbname, "_design/foo", "foo.txt")
		})
		c.RunNoAuth(t, func(t *testing.T) {
			t.Helper()
			t.Parallel()
			testGetAttachments(t, c, c.NoAuth, dbname, "foo", "foo.txt")
			testGetAttachments(t, c, c.NoAuth, dbname, "foo", "NotFound")
			testGetAttachments(t, c, c.NoAuth, dbname, "_design/foo", "foo.txt")
		})
	})
}

func testGetAttachments(t *testing.T, c *kt.Context, client *kivik.Client, dbname, docID, filename string) { //nolint:thelper
	c.Run(t, docID+"/"+filename, func(t *testing.T) {
		t.Parallel()
		db := c.DB(t, client, dbname)
		att, err := db.GetAttachment(context.Background(), docID, filename)
		if !c.IsExpectedSuccess(t, err) {
			return
		}
		if client.Driver() != "pouch" {
			if att.ContentType != "text/plain" {
				t.Errorf("Content-Type: Expected %s, Actual %s", "text/plain", att.ContentType)
			}
		}
	})
}
