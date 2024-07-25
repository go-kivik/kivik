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

package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/mockdb"
)

func Test_clusterStatus(t *testing.T) {
	tests := serverTests{
		{
			name:       "cluster status, unauthorized",
			method:     http.MethodGet,
			path:       "/_cluster_setup",
			wantStatus: http.StatusUnauthorized,
			wantJSON: map[string]interface{}{
				"error":  "unauthorized",
				"reason": "User not authenticated",
			},
		},
		{
			name: "cluster status, success",
			client: func() *kivik.Client {
				client, mock, err := mockdb.New()
				if err != nil {
					t.Fatal(err)
				}
				mock.ExpectClusterStatus().
					WillReturn("chicken")
				return client
			}(),
			method:     http.MethodGet,
			path:       "/_cluster_setup",
			authUser:   userAdmin,
			wantStatus: http.StatusOK,
			wantJSON: map[string]string{
				"state": "chicken",
			},
		},
	}

	tests.Run(t)
}

func TestClusterSetup(t *testing.T) {
	tests := serverTests{
		{
			name:       "cluster status, unauthorized",
			method:     http.MethodPost,
			path:       "/_cluster_setup",
			wantStatus: http.StatusUnauthorized,
			wantJSON: map[string]string{
				"error":  "unauthorized",
				"reason": "User not authenticated",
			},
		},
		{
			name: "cluster status, success",
			client: func() *kivik.Client {
				client, mock, err := mockdb.New()
				if err != nil {
					t.Fatal(err)
				}
				mock.ExpectClusterSetup().
					WithAction("chicken").
					WillReturnError(nil)
				return client
			}(),
			method:     http.MethodPost,
			authUser:   userAdmin,
			path:       "/_cluster_setup",
			headers:    map[string]string{"Content-Type": "application/json"},
			body:       strings.NewReader(`{"action":"chicken"}`),
			wantStatus: http.StatusOK,
			wantJSON: map[string]bool{
				"ok": true,
			},
		},
	}

	tests.Run(t)
}
