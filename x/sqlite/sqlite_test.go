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
