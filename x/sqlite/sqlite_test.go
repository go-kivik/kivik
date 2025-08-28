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
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

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

	ver, err := c.Version(context.Background())
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

func TestClientAllDBs(t *testing.T) {
	d := drv{}
	dClient, err := d.NewClient(":memory:", mock.NilOption)
	if err != nil {
		t.Fatal(err)
	}

	c := dClient.(*client)
	for _, table := range []string{"foo", "bar", "foo_attachments", "bar_attachments", "foo_attachments_bridge", "foo_design", "bar_design", "foo_revs", "bar_revs"} {
		if _, err := c.db.Exec("CREATE TABLE " + table + " (id INTEGER)"); err != nil {
			t.Fatal(err)
		}
	}

	dbs, err := dClient.AllDBs(context.Background(), mock.NilOption)
	if err != nil {
		t.Fatal("err should be nil")
	}
	wantDBs := []string{"foo", "bar"}
	if d := cmp.Diff(wantDBs, dbs); d != "" {
		t.Fatal(d)
	}
}

func TestClientDBExists(t *testing.T) {
	d := drv{}
	t.Run("exists", func(t *testing.T) {
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		c := dClient.(*client)

		if _, err := c.db.Exec("CREATE TABLE foo (id INTEGER)"); err != nil {
			t.Fatal(err)
		}

		exists, err := dClient.DBExists(context.Background(), "foo", nil)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatal("foo should exist")
		}
	})
	t.Run("does not exist", func(t *testing.T) {
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		exists, err := dClient.DBExists(context.Background(), "foo", nil)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("foo should not exist")
		}
	})
}

func TestClientCreateDB(t *testing.T) {
	t.Run("invalid name", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		err = dClient.CreateDB(context.Background(), "Foo", mock.NilOption)
		if err == nil {
			t.Fatal("err should not be nil")
		}
		const wantErr = "invalid database name: Foo"
		if !testy.ErrorMatches(wantErr, err) {
			t.Fatalf("Unexpected error: %s", err)
		}
		const wantStatus = http.StatusBadRequest
		if status := kivik.HTTPStatus(err); status != wantStatus {
			t.Fatalf("status should be %d", wantStatus)
		}
	})
	t.Run("success", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		if err := dClient.CreateDB(context.Background(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		exists, err := dClient.DBExists(context.Background(), "foo", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatal("foo should exist")
		}
	})
	t.Run("db already exists", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		if err := dClient.CreateDB(context.Background(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		err = dClient.CreateDB(context.Background(), "foo", mock.NilOption)
		if err == nil {
			t.Fatal("err should not be nil")
		}
		const wantErr = "database already exists"
		if err.Error() != wantErr {
			t.Fatalf("err should be %s, got %s", wantErr, err)
		}
		const wantStatus = http.StatusPreconditionFailed
		if status := kivik.HTTPStatus(err); status != wantStatus {
			t.Fatalf("status should be %d", wantStatus)
		}
	})
}

func TestClientDestroyDB(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		if err := dClient.CreateDB(context.Background(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		if err := dClient.DestroyDB(context.Background(), "foo", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		exists, err := dClient.DBExists(context.Background(), "foo", mock.NilOption)
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

		err = dClient.DestroyDB(context.Background(), "foo", nil)
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
