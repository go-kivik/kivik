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
	"encoding/json"
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
	// TODO: update in Cycle 5 when kivik$db_updates_log is removed
	t.Skip("Skipped until Cycle 5 when kivik$db_updates_log is removed")
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
		dClient := testClient(t)

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
			options: mock.NilOption,
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
		dClient := testClient(t)

		return test{
			client:     dClient.(*client),
			options:    kivik.Param("feed", "invalid"),
			wantStatus: 400,
			wantErr:    "supported `feed` types: normal, longpoll, continuous",
		}
	})

	tests.Add("database deletion events are logged", func(t *testing.T) interface{} {
		dClient := testClient(t)

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
			options: mock.NilOption,
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
		dClient := testClient(t)

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
	// TODO: update in Cycle 5 when kivik$db_updates_log is removed
	t.Skip("Skipped until Cycle 5 when kivik$db_updates_log is removed")
	t.Parallel()

	dClient := testClient(t).(*client)

	ctx := context.Background()
	if err := dClient.CreateDB(ctx, "db1", mock.NilOption); err != nil {
		t.Fatal(err)
	}

	updates, err := dClient.DBUpdates(ctx, kivik.Params(map[string]any{
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

func TestClientDBUpdates_globalChanges(t *testing.T) {
	t.Parallel()

	type test struct {
		client     *client
		options    driver.Options
		wantStatus int
		wantErr    string
	}

	tests := testy.NewTable()

	tests.Add("returns 503 when _global_changes is absent", func(t *testing.T) interface{} {
		dClient := testClient(t)

		return test{
			client:     dClient.(*client),
			options:    mock.NilOption,
			wantStatus: 503,
			wantErr:    "Service Unavailable",
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		_, err := tt.client.DBUpdates(context.Background(), tt.options)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("unexpected error, got %s, want %s", err, tt.wantErr)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: got %d, want %d", status, tt.wantStatus)
		}
	})
}

func TestClientLogGlobalChange(t *testing.T) {
	t.Parallel()

	type changeDoc struct {
		DBName string `json:"db_name"`
		Type   string `json:"type"`
	}

	type test struct {
		dClient     driver.Client
		wantChanges []changeDoc
		wantErr     string
	}

	newSingleNodeClient := func(t *testing.T) driver.Client {
		t.Helper()
		dClient := testClient(t)
		cluster := dClient.(driver.Cluster)
		if err := cluster.ClusterSetup(context.Background(), map[string]any{"action": "enable_single_node"}); err != nil {
			t.Fatal(err)
		}
		return dClient
	}

	tests := testy.NewTable()

	tests.Add("creating a db after enable_single_node logs a change to _global_changes", func(t *testing.T) interface{} {
		dClient := newSingleNodeClient(t)
		if err := dClient.CreateDB(context.Background(), "testdb", mock.NilOption); err != nil {
			t.Fatal(err)
		}
		return test{
			dClient:     dClient,
			wantChanges: []changeDoc{{DBName: "testdb", Type: "created"}},
		}
	})

	tests.Add("destroying a db after enable_single_node logs a deleted change to _global_changes", func(t *testing.T) interface{} {
		dClient := newSingleNodeClient(t)
		ctx := context.Background()
		if err := dClient.CreateDB(ctx, "testdb", mock.NilOption); err != nil {
			t.Fatal(err)
		}
		if err := dClient.DestroyDB(ctx, "testdb", mock.NilOption); err != nil {
			t.Fatal(err)
		}
		return test{
			dClient: dClient,
			wantChanges: []changeDoc{
				{DBName: "testdb", Type: "created"},
				{DBName: "testdb", Type: "deleted"},
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		ctx := context.Background()

		globalDB, err := tt.dClient.DB("_global_changes", mock.NilOption)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("unexpected error opening _global_changes, got %s, want %s", err, tt.wantErr)
		}
		if err != nil {
			return
		}
		t.Cleanup(func() { _ = globalDB.Close() })

		feed, err := globalDB.Changes(ctx, kivik.Param("include_docs", true))
		if err != nil {
			t.Fatalf("unexpected error starting changes feed: %s", err)
		}
		t.Cleanup(func() { _ = feed.Close() })

		var got []changeDoc
		for {
			var change driver.Change
			if err := feed.Next(&change); err != nil {
				break
			}
			var doc changeDoc
			if err := json.Unmarshal(change.Doc, &doc); err != nil {
				t.Fatalf("failed to unmarshal doc: %s", err)
			}
			got = append(got, doc)
		}

		if d := cmp.Diff(tt.wantChanges, got); d != "" {
			t.Errorf("Unexpected changes in _global_changes: %s", d)
		}
	})
}
