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

package sqlite

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/go-kivik/kivik/v4/driver"
)

func TestNewClient(t *testing.T) {
	d := drv{}

	client, _ := d.NewClient("xxx", nil)
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
	dClient, err := d.NewClient(":memory:", nil)
	if err != nil {
		t.Fatal(err)
	}

	c := dClient.(*client)

	if _, err := c.db.Exec("CREATE TABLE foo (id INTEGER)"); err != nil {
		t.Fatal(err)
	}
	if _, err := c.db.Exec("CREATE TABLE bar (id INTEGER)"); err != nil {
		t.Fatal(err)
	}

	dbs, err := c.AllDBs(context.Background(), nil)
	if err != nil {
		t.Fatal("err should be nil")
	}
	wantDBs := []string{"foo", "bar"}
	if d := cmp.Diff(wantDBs, dbs); d != "" {
		t.Fatal(d)
	}
}
