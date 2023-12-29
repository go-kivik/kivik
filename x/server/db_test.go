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
// +build !js

package server

import (
	"net/http"
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/mockdb"
)

func Test_dbUpdates(t *testing.T) {
	tests := serverTests{
		{
			name:       "db updates, unauthorized",
			method:     http.MethodGet,
			path:       "/_db_updates",
			wantStatus: http.StatusUnauthorized,
			wantJSON: map[string]interface{}{
				"error":  "unauthorized",
				"reason": "User not authenticated",
			},
		},
		{
			name: "db updates, two updates",
			client: func() *kivik.Client {
				client, mock, err := mockdb.New()
				if err != nil {
					t.Fatal(err)
				}
				mock.ExpectDBUpdates().WillReturn(mockdb.NewDBUpdates().
					AddUpdate(&driver.DBUpdate{
						DBName: "foo",
						Type:   "created",
						Seq:    "1-aaa",
					}).
					AddUpdate(&driver.DBUpdate{
						DBName: "foo",
						Type:   "deleted",
						Seq:    "2-aaa",
					}))
				return client
			}(),
			authUser:   userAdmin,
			method:     http.MethodGet,
			path:       "/_db_updates",
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"results": []interface{}{
					map[string]interface{}{
						"db_name": "foo",
						"type":    "created",
						"seq":     "1-aaa",
					},
					map[string]interface{}{
						"db_name": "foo",
						"type":    "deleted",
						"seq":     "2-aaa",
					},
				},
				"last_seq": "2-aaa",
			},
		},
		{
			name:       "db updates, invalid heartbeat",
			method:     http.MethodGet,
			authUser:   userAdmin,
			path:       "/_db_updates?heartbeat=chicken",
			wantStatus: http.StatusBadRequest,
			wantJSON: map[string]interface{}{
				"error":  "bad_request",
				"reason": "strconv.Atoi: parsing \"chicken\": invalid syntax",
			},
		},
	}

	tests.Run(t)
}
