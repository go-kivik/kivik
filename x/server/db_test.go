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
	"strings"
	"testing"
	"time"

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
					}).LastSeq("2-aaa"))
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
			name:       "continuous, invalid heartbeat",
			method:     http.MethodGet,
			authUser:   userAdmin,
			path:       "/_db_updates?feed=continuous&heartbeat=chicken",
			wantStatus: http.StatusBadRequest,
			wantJSON: map[string]interface{}{
				"error":  "bad_request",
				"reason": "strconv.Atoi: parsing \"chicken\": invalid syntax",
			},
		},
		{
			name: "continuous, with heartbeat",
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
					AddDelay(500 * time.Millisecond).
					AddUpdate(&driver.DBUpdate{
						DBName: "foo",
						Type:   "deleted",
						Seq:    "2-aaa",
					}))
				return client
			}(),
			authUser:   userAdmin,
			method:     http.MethodGet,
			path:       "/_db_updates?feed=continuous&heartbeat=100",
			wantStatus: http.StatusOK,
			wantBodyRE: "}\n+\n{",
		},
	}

	tests.Run(t)
}

func Test_allDocs(t *testing.T) {
	tests := serverTests{
		{
			name:     "GET defaults",
			authUser: userAdmin,
			method:   http.MethodGet,
			path:     "/db1/_all_docs",
			client: func() *kivik.Client {
				client, mock, err := mockdb.New()
				if err != nil {
					t.Fatal(err)
				}
				db := mock.NewDB()
				mock.ExpectDB().WillReturn(db)
				db.ExpectSecurity().WillReturn(&driver.Security{})
				mock.ExpectDB().WillReturn(db)
				db.ExpectAllDocs().WillReturn(mockdb.NewRows().
					AddRow(&driver.Row{
						ID:    "foo",
						Key:   []byte(`"foo"`),
						Value: strings.NewReader(`{"rev": "1-beea34a62a215ab051862d1e5d93162e"}`),
					}).
					TotalRows(99),
				)
				return client
			}(),
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"offset": 0,
				"rows": []interface{}{
					map[string]interface{}{
						"id":  "foo",
						"key": "foo",
						"value": map[string]interface{}{
							"rev": "1-beea34a62a215ab051862d1e5d93162e",
						},
					},
				},
				"total_rows": 99,
			},
		},
		{
			name:     "multi queries defaults",
			authUser: userAdmin,
			method:   http.MethodPost,
			path:     "/db1/_all_docs/queries",
			headers: map[string]string{
				"Content-Type": "application/json",
			},
			body: strings.NewReader(`{"queries": [{"keys": ["foo", "bar"]}]}`),
			client: func() *kivik.Client {
				client, mock, err := mockdb.New()
				if err != nil {
					t.Fatal(err)
				}
				db := mock.NewDB()
				mock.ExpectDB().WillReturn(db)
				db.ExpectSecurity().WillReturn(&driver.Security{})
				mock.ExpectDB().WillReturn(db)
				db.ExpectAllDocs().WillReturn(mockdb.NewRows().
					AddRow(&driver.Row{
						ID:    "foo",
						Key:   []byte(`"foo"`),
						Value: strings.NewReader(`{"rev": "1-beea34a62a215ab051862d1e5d93162e"}`),
					}).
					TotalRows(99),
				)
				return client
			}(),
			wantStatus: http.StatusOK,
			wantJSON: map[string]interface{}{
				"offset": 0,
				"rows": []interface{}{
					map[string]interface{}{
						"id":  "foo",
						"key": "foo",
						"value": map[string]interface{}{
							"rev": "1-beea34a62a215ab051862d1e5d93162e",
						},
					},
				},
				"total_rows": 99,
			},
		},
	}

	tests.Run(t)
}
