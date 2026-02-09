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

	"github.com/go-kivik/kivik/v4/driver"
)

func TestSecurity(t *testing.T) {
	t.Parallel()

	type test struct {
		security *driver.Security
		want     *driver.Security
		wantErr  string
	}

	tests := testy.NewTable()

	tests.Add("round-trip with admins and members", test{
		security: &driver.Security{
			Admins: driver.Members{
				Names: []string{"alice"},
				Roles: []string{"admin"},
			},
			Members: driver.Members{
				Names: []string{"bob"},
				Roles: []string{"reader"},
			},
		},
		want: &driver.Security{
			Admins: driver.Members{
				Names: []string{"alice"},
				Roles: []string{"admin"},
			},
			Members: driver.Members{
				Names: []string{"bob"},
				Roles: []string{"reader"},
			},
		},
	})

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		d := newDB(t)
		secDB := d.DB.(driver.SecurityDB)

		err := secDB.SetSecurity(context.Background(), tt.security)
		if !testy.ErrorMatchesRE(tt.wantErr, err) {
			t.Fatalf("SetSecurity: unexpected error, got %s, want /%s/", err, tt.wantErr)
		}
		if err != nil {
			return
		}

		got, err := secDB.Security(context.Background())
		if err != nil {
			t.Fatalf("Security: unexpected error: %s", err)
		}
		if d := cmp.Diff(tt.want, got); d != "" {
			t.Errorf("Unexpected security document:\n%s", d)
		}
	})
}
