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
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestClientDBUpdates(t *testing.T) {
	t.Parallel()

	type test struct {
		client     *client
		options    driver.Options
		want       []driver.DBUpdate
		wantStatus int
		wantErr    string
	}

	tests := testy.NewTable()

	tests.Add("no since parameter returns all events", func(t *testing.T) interface{} {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		ctx := context.Background()
		if err := dClient.CreateDB(ctx, "db1", mock.NilOption); err != nil {
			t.Fatal(err)
		}
		if err := dClient.CreateDB(ctx, "db2", mock.NilOption); err != nil {
			t.Fatal(err)
		}
		if err := dClient.CreateDB(ctx, "db3", mock.NilOption); err != nil {
			t.Fatal(err)
		}

		return test{
			client:  dClient.(*client),
			options: nil,
			want: []driver.DBUpdate{
				{
					DBName: "db1",
					Type:   "created",
					Seq:    "1",
				},
				{
					DBName: "db2",
					Type:   "created",
					Seq:    "2",
				},
				{
					DBName: "db3",
					Type:   "created",
					Seq:    "3",
				},
			},
		}
	})

	tests.Add("invalid feed value is rejected", func(t *testing.T) interface{} {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		return test{
			client:     dClient.(*client),
			options:    kivik.Param("feed", "invalid"),
			wantStatus: 400,
			wantErr:    "supported `feed` types: normal, longpoll, continuous",
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		ctx := context.Background()
		updates, err := tt.client.DBUpdates(ctx, tt.options)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("unexpected error, got %s, want %s", err, tt.wantErr)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		defer updates.Close()

		var got []driver.DBUpdate
		for {
			var update driver.DBUpdate
			if err := updates.Next(&update); err != nil {
				break
			}
			got = append(got, update)
		}

		if d := cmp.Diff(tt.want, got); d != "" {
			t.Errorf("Unexpected result: %s", d)
		}
	})
}
