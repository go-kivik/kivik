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

package fs

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

func Test_db_Stats(t *testing.T) {
	tests := []struct {
		name         string
		path, dbname string
		wantStatus   int
		wantErr      string
		want         *driver.DBStats
	}{
		{
			name:   "success",
			path:   "testdata",
			dbname: "db_att",
			want: &driver.DBStats{
				Name: "db_att",
			},
		},
		{
			name:       "not found",
			path:       "testdata",
			dbname:     "notfound",
			wantStatus: http.StatusNotFound,
			wantErr:    `[Nn]o such file or directory`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &client{root: tt.path, fs: filesystem.Default()}
			db, err := c.newDB(tt.dbname)
			if err != nil {
				t.Fatal(err)
			}

			stats, err := db.Stats(t.Context())
			if !testy.ErrorMatchesRE(tt.wantErr, err) {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if status := kivik.HTTPStatus(err); status != tt.wantStatus {
				t.Errorf("status = %v, wantStatus %v", status, tt.wantStatus)
			}
			if d := cmp.Diff(tt.want, stats); d != "" {
				t.Errorf("Unexpected stats:\n%s\n", d)
			}
		})
	}
}
