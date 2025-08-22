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

package pg

import (
	"net/http"
	"testing"

	"github.com/go-kivik/kivik/v4"
)

func TestNewClient(t *testing.T) {
	t.Parallel()

	type test struct {
		name       string
		dsn        string
		wantErr    bool
		wantStatus int
	}
	tests := []test{
		{
			name:       "invalid dsn",
			dsn:        "completely invalid dsn",
			wantErr:    true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "valid dsn with missing dbname",
			dsn:        "postgres://user:pass@localhost:5432",
			wantErr:    true,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:    "valid dsn with dbname",
			dsn:     "postgres://user:pass@localhost:5432/dbname",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			_, err := (&pg{}).NewClient(tt.dsn, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient(%q) error = %v, wantErr %v", tt.dsn, err, tt.wantErr)
			}
			if status := kivik.HTTPStatus(err); status != tt.wantStatus {
				t.Errorf("NewClient(%q) status = %d, want %d", tt.dsn, status, tt.wantStatus)
			}
		})
	}
}
