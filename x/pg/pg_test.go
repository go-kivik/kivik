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
		dsn string
		err bool
	}
	tests := []test{
		{
			dsn: "completely invalid dsn",
			err: true,
		},
	}

	/*
		TODO:
		- DSN does not include a database name -- should fail
		- DSN as URL
		- DSN as key/value pairs
	*/

	for _, tt := range tests {
		t.Run(tt.dsn, func(t *testing.T) {
			t.Parallel()

			_, err := (&pg{}).NewClient(tt.dsn, nil)
			if (err != nil) != tt.err {
				t.Errorf("NewClient(%q) error = %v, wantErr %v", tt.dsn, err, tt.err)
			}
			if err != nil {
				return
			}
			status := kivik.HTTPStatus(err)
			if status != http.StatusBadRequest {
				t.Errorf("NewClient(%q) status = %d, want %d", tt.dsn, status, http.StatusBadRequest)
			}
		})
	}
}
