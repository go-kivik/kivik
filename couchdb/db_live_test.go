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

package couchdb_test

import (
	"context"
	"os"
	"runtime"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
)

const isGopherJS = runtime.GOOS == "js" || runtime.GOARCH == "js"

func TestQueries_2_x(t *testing.T) {
	if isGopherJS {
		t.Skip("Network tests skipped for GopherJS")
	}
	dsn := os.Getenv("KIVIK_TEST_DSN_COUCH23")
	if dsn == "" {
		dsn = os.Getenv("KIVIK_TEST_DSN_COUCH22")
	}
	if dsn == "" {
		t.Skip("Neither KIVIK_TEST_DSN_COUCH22 nor KIVIK_TEST_DSN_COUCH23 configured")
	}

	client, err := kivik.New("couch", dsn)
	if err != nil {
		t.Fatal(err)
	}

	db := client.DB("_users")
	rows := db.AllDocs(context.Background(), kivik.Params(map[string]any{
		"queries": []map[string]any{
			{},
			{},
		},
	}))
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = rows.Close()
	})
	result := make([]any, 0)
	for rows.Next() {
		id, _ := rows.ID()
		result = append(result, map[string]any{
			"_id": id,
		})
	}
	meta, err := rows.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	wantMeta := &kivik.ResultMetadata{
		TotalRows: 1,
	}
	if d := testy.DiffInterface(wantMeta, meta); d != nil {
		t.Error(d)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if d := testy.DiffInterface(testy.Snapshot(t), result); d != nil {
		t.Error(d)
	}
}

func TestQueries_3_x(t *testing.T) {
	dsn := os.Getenv("KIVIK_TEST_DSN_COUCH30")
	if dsn == "" {
		t.Skip("KIVIK_TEST_DSN_COUCH30 not configured")
	}

	client, err := kivik.New("couch", dsn)
	if err != nil {
		t.Fatal(err)
	}

	db := client.DB("_users")
	rows := db.AllDocs(context.Background(), kivik.Params(map[string]any{
		"queries": []map[string]any{
			{},
			{},
		},
	}))
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = rows.Close()
	})
	result := make([]any, 0)
	for rows.Next() {
		id, _ := rows.ID()
		result = append(result, map[string]any{
			"_id": id,
		})
	}
	meta, err := rows.Metadata()
	if err != nil {
		t.Fatal(err)
	}
	wantMeta := &kivik.ResultMetadata{
		TotalRows: 1,
	}
	if d := testy.DiffInterface(wantMeta, meta); d != nil {
		t.Error(d)
	}
	if err := rows.Err(); err != nil {
		t.Fatal(err)
	}
	if d := testy.DiffInterface(testy.Snapshot(t), result); d != nil {
		t.Error(d)
	}
}

// https://github.com/go-kivik/kivik/issues/509
func Test_bug509(t *testing.T) {
	if isGopherJS {
		t.Skip("Network tests skipped for GopherJS")
	}
	dsn := os.Getenv("KIVIK_TEST_DSN_COUCH23")
	if dsn == "" {
		t.Skip("KIVIK_TEST_DSN_COUCH23 not configured")
	}

	client, err := kivik.New("couch", dsn)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = client.DestroyDB(context.Background(), "bug509")
		_ = client.Close()
	})
	if err := client.CreateDB(context.Background(), "bug509"); err != nil {
		t.Fatal(err)
	}

	db := client.DB("bug509")
	if _, err := db.Put(context.Background(), "x", map[string]string{
		"_id": "x",
	}); err != nil {
		t.Fatal(err)
	}
	row := db.Get(context.Background(), "x", kivik.Param("revs_info", true))

	var doc map[string]any
	if err := row.ScanDoc(&doc); err != nil {
		t.Fatal(err)
	}
}
