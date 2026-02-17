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
	"sort"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestClientClusterSetup(t *testing.T) {
	t.Parallel()

	type test struct {
		action  any
		want    []string
		wantErr string
	}

	tests := testy.NewTable()

	tests.Add("enable_single_node creates system databases", test{
		action: map[string]any{"action": "enable_single_node"},
		want:   []string{"_global_changes", "_replicator", "_users"},
	})
	tests.Add("finish_cluster creates system databases", test{
		action: map[string]any{"action": "finish_cluster"},
		want:   []string{"_global_changes", "_replicator", "_users"},
	})

	tests.Run(t, func(t *testing.T, tt test) {
		d := drv{}
		dClient, err := d.NewClient(":memory:", mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}

		c := dClient.(driver.Cluster)

		err = c.ClusterSetup(context.Background(), tt.action)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("unexpected error, got %s, want %s", err, tt.wantErr)
		}
		if err != nil {
			return
		}

		got, err := dClient.AllDBs(context.Background(), mock.NilOption)
		if err != nil {
			t.Fatal(err)
		}
		sort.Strings(got)

		if d := cmp.Diff(tt.want, got); d != "" {
			t.Errorf("Unexpected databases: %s", d)
		}
	})
}
