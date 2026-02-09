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
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestClientAllDBs(t *testing.T) {
	t.Parallel()
	type test struct {
		client  *client
		options driver.Options
		want    []string
	}

	tests := testy.NewTable()

	tests.Add("returns databases in ascending order", func(t *testing.T) interface{} {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		c := dClient.(*client)
		for _, table := range []string{
			"kivik$foo", "kivik$bar",
			"kivik$foo$attachments", "kivik$bar$attachments",
			"kivik$foo$attachments_bridge", "kivik$bar$attachments_bridge",
			"kivik$foo$design", "kivik$bar$design",
			"kivik$foo$revs", "kivik$bar$revs",
			"kivik$testddoc_1-abc_myview_map_12345678",
		} {
			if _, err := c.db.Exec("CREATE TABLE \"" + table + "\" (id INTEGER)"); err != nil {
				t.Fatal(err)
			}
		}
		return test{
			client: c,
			want:   []string{"bar", "foo"},
		}
	})
	tests.Add("descending=true", func(t *testing.T) interface{} {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		c := dClient.(*client)
		ctx := context.Background()
		for _, name := range []string{"aaa", "bbb", "ccc"} {
			if err := c.CreateDB(ctx, name, mock.NilOption); err != nil {
				t.Fatal(err)
			}
		}
		return test{
			client:  c,
			options: kivik.Param("descending", true),
			want:    []string{"ccc", "bbb", "aaa"},
		}
	})
	tests.Add("limit=2", func(t *testing.T) interface{} {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		c := dClient.(*client)
		ctx := context.Background()
		for _, name := range []string{"aaa", "bbb", "ccc"} {
			if err := c.CreateDB(ctx, name, mock.NilOption); err != nil {
				t.Fatal(err)
			}
		}
		return test{
			client:  c,
			options: kivik.Param("limit", 2),
			want:    []string{"aaa", "bbb"},
		}
	})
	tests.Add("skip=1", func(t *testing.T) interface{} {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		c := dClient.(*client)
		ctx := context.Background()
		for _, name := range []string{"aaa", "bbb", "ccc"} {
			if err := c.CreateDB(ctx, name, mock.NilOption); err != nil {
				t.Fatal(err)
			}
		}
		return test{
			client:  c,
			options: kivik.Param("skip", 1),
			want:    []string{"bbb", "ccc"},
		}
	})
	tests.Add("startkey", func(t *testing.T) interface{} {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		c := dClient.(*client)
		ctx := context.Background()
		for _, name := range []string{"aaa", "bbb", "ccc"} {
			if err := c.CreateDB(ctx, name, mock.NilOption); err != nil {
				t.Fatal(err)
			}
		}
		return test{
			client:  c,
			options: kivik.Param("startkey", "bbb"),
			want:    []string{"bbb", "ccc"},
		}
	})
	tests.Add("endkey", func(t *testing.T) interface{} {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		c := dClient.(*client)
		ctx := context.Background()
		for _, name := range []string{"aaa", "bbb", "ccc"} {
			if err := c.CreateDB(ctx, name, mock.NilOption); err != nil {
				t.Fatal(err)
			}
		}
		return test{
			client:  c,
			options: kivik.Param("endkey", "bbb"),
			want:    []string{"aaa", "bbb"},
		}
	})
	tests.Add("inclusive_end=false", func(t *testing.T) interface{} {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		c := dClient.(*client)
		ctx := context.Background()
		for _, name := range []string{"aaa", "bbb", "ccc"} {
			if err := c.CreateDB(ctx, name, mock.NilOption); err != nil {
				t.Fatal(err)
			}
		}
		return test{
			client: c,
			options: kivik.Params(map[string]interface{}{
				"endkey":        "bbb",
				"inclusive_end": false,
			}),
			want: []string{"aaa"},
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}

		dbs, err := tt.client.AllDBs(context.Background(), opts)
		if err != nil {
			t.Fatal(err)
		}
		if d := cmp.Diff(tt.want, dbs); d != "" {
			t.Errorf("Unexpected result:\n%s", d)
		}
	})
}
