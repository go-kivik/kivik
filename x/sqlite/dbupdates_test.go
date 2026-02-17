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
	"fmt"
	"sync"
	"testing"
	"time"

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

	tests.Add("database deletion events are logged", func(t *testing.T) interface{} {
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
		if err := dClient.DestroyDB(ctx, "db1", mock.NilOption); err != nil {
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
					DBName: "db1",
					Type:   "deleted",
					Seq:    "3",
				},
			},
		}
	})

	tests.Add("string sequence value in since parameter", func(t *testing.T) interface{} {
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
			options: kivik.Param("since", "1-foo"),
			want: []driver.DBUpdate{
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

func TestClientDBUpdates_longpoll(t *testing.T) {
	t.Parallel()

	d := drv{}
	driverClient, err := d.NewClient(":memory:", mock.NilOption)
	if err != nil {
		t.Fatal(err)
	}
	dClient := driverClient.(*client)

	ctx := context.Background()
	if err := dClient.CreateDB(ctx, "db1", mock.NilOption); err != nil {
		t.Fatal(err)
	}

	updates, err := dClient.DBUpdates(context.Background(), kivik.Params(map[string]any{
		"feed":  "longpoll",
		"since": "now",
	}))
	if err != nil {
		t.Fatalf("Failed to start DBUpdates feed: %s", err)
	}
	t.Cleanup(func() {
		_ = updates.Close()
	})

	var mu sync.Mutex
	var createdDB string

	go func() {
		time.Sleep(100 * time.Millisecond)
		if err := dClient.CreateDB(context.Background(), "db2", mock.NilOption); err != nil {
			panic(fmt.Sprintf("Failed to create db: %s", err))
		}
		mu.Lock()
		createdDB = "db2"
		mu.Unlock()
	}()

	start := time.Now()

	var update driver.DBUpdate
	if err := updates.Next(&update); err != nil {
		t.Fatalf("Failed to get next update: %s", err)
	}

	elapsed := time.Since(start)
	if elapsed < 50*time.Millisecond {
		t.Errorf("Expected feed to block, but returned in %v", elapsed)
	}

	mu.Lock()
	defer mu.Unlock()

	want := driver.DBUpdate{
		DBName: createdDB,
		Type:   "created",
		Seq:    "2",
	}

	if d := cmp.Diff(want, update); d != "" {
		t.Errorf("Unexpected update: %s", d)
	}
}
