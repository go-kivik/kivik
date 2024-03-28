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
	"os"
	"testing"

	"github.com/go-kivik/kivik/v4/driver"
)

// newDB creates a new driver.DB instance backed by an in-memory SQLite database,
// and registers a cleanup function to close the database when the test is done.
func newDB(t *testing.T) driver.DB {
	dsn := ":memory:"
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
	return db
}
