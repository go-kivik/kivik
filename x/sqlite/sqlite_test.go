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

package sqlite

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestNewClient(t *testing.T) {
	d := drv{}

	client, _ := d.NewClient(":memory:", mock.NilOption)
	if client == nil {
		t.Fatal("client should not be nil")
	}
}

func TestClientVersion(t *testing.T) {
	c := client{}

	ver, err := c.Version(t.Context())
	if err != nil {
		t.Fatal("err should be nil")
	}
	wantVer := &driver.Version{
		Version: "0.0.1",
		Vendor:  "Kivik",
	}
	if d := cmp.Diff(wantVer, ver); d != "" {
		t.Fatal(d)
	}
}

func TestConcurrentPuts(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "kivik-concurrent-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dsn := f.Name() + "?_txlock=immediate&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	_ = f.Close()
	t.Cleanup(func() { _ = os.Remove(dsn) })

	d := drv{}
	dClient, err := d.NewClient(dsn, mock.NilOption)
	if err != nil {
		t.Fatal(err)
	}

	if err := dClient.CreateDB(t.Context(), "test", mock.NilOption); err != nil {
		t.Fatal(err)
	}
	db, err := dClient.DB("test", mock.NilOption)
	if err != nil {
		t.Fatal(err)
	}

	errs := make(chan error, 5)
	for i := range 5 {
		go func() {
			_, err := db.Put(t.Context(), fmt.Sprintf("doc%d", i), map[string]string{"key": "val"}, mock.NilOption)
			if err != nil {
				t.Logf("g%d error: %v", i, err)
			}
			errs <- err
		}()
	}
	for range 5 {
		if err := <-errs; err != nil {
			t.Errorf("concurrent put failed: %s", err)
		}
	}
}

func TestMultipleDBs(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "kivik-multi-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dsn := f.Name()
	_ = f.Close()
	t.Cleanup(func() { _ = os.Remove(dsn) })

	d := drv{}
	dClient, err := d.NewClient(dsn, mock.NilOption)
	if err != nil {
		t.Fatal(err)
	}
	if err := dClient.CreateDB(t.Context(), "db1", mock.NilOption); err != nil {
		t.Fatalf("creating db1: %v", err)
	}
	if err := dClient.CreateDB(t.Context(), "db2", mock.NilOption); err != nil {
		t.Fatalf("creating db2: %v", err)
	}
}

func TestClientDestroyDB(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		if err := dClient.CreateDB(t.Context(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		if err := dClient.DestroyDB(t.Context(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		exists, err := dClient.DBExists(t.Context(), "foo", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("foo should not exist")
		}
	})
	t.Run("with design doc", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		if err := dClient.CreateDB(t.Context(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		db, err := dClient.DB("foo", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		ddoc := map[string]interface{}{
			"_id":      "_design/testddoc",
			"language": "javascript",
			"views": map[string]interface{}{
				"testview": map[string]interface{}{
					"map": `function(doc) { emit(doc._id, null); }`,
				},
			},
		}
		if _, err := db.Put(t.Context(), "_design/testddoc", ddoc, mock.NilOption); err != nil {
			t.Fatal(err)
		}

		if err := dClient.DestroyDB(t.Context(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		exists, err := dClient.DBExists(t.Context(), "foo", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("foo should not exist")
		}
	})
	t.Run("doesn't exist", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		err = dClient.DestroyDB(t.Context(), "foo", nil)
		if err == nil {
			t.Fatal("wanted an error")
		}
		const wantErr = "database not found"
		if err.Error() != wantErr {
			t.Fatalf("err should be %s, got %s", wantErr, err)
		}
		const wantStatus = http.StatusNotFound
		if status := kivik.HTTPStatus(err); status != wantStatus {
			t.Fatalf("status should be %d", wantStatus)
		}
	})
}
