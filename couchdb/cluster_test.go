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

package couchdb

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

const optionEnsureDBsExist = "ensure_dbs_exist"

func TestClusterStatus(t *testing.T) {
	type tst struct {
		client   *client
		options  map[string]interface{}
		expected string
		status   int
		err      string
	}
	tests := testy.NewTable()
	tests.Add("network error", tst{
		client: newTestClient(nil, errors.New("network error")),
		status: http.StatusBadGateway,
		err:    `Get "?http://example.com/_cluster_setup"?: network error`,
	})
	tests.Add("finished", tst{
		client: newTestClient(&http.Response{
			StatusCode: http.StatusOK,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`{"state":"cluster_finished"}`)),
		}, nil),
		expected: "cluster_finished",
	})
	tests.Add("invalid option", tst{
		client: newCustomClient(func(r *http.Request) (*http.Response, error) {
			return nil, nil
		}),
		options: map[string]interface{}{
			optionEnsureDBsExist: 1.0,
		},
		status: http.StatusBadRequest,
		err:    "kivik: invalid type float64 for options",
	})
	tests.Add("invalid param", tst{
		client: newCustomClient(func(r *http.Request) (*http.Response, error) {
			result := []string{}
			err := json.Unmarshal([]byte(r.URL.Query().Get(optionEnsureDBsExist)), &result)
			return nil, &kivik.Error{Status: http.StatusBadRequest, Err: err}
		}),
		options: map[string]interface{}{
			optionEnsureDBsExist: "foo,bar,baz",
		},
		status: http.StatusBadRequest,
		err:    `Get "?http://example.com/_cluster_setup\?ensure_dbs_exist=foo%2Cbar%2Cbaz"?: invalid character 'o' in literal false \(expecting 'a'\)`,
	})
	tests.Add("ensure dbs", func(t *testing.T) interface{} {
		return tst{
			client: newCustomClient(func(r *http.Request) (*http.Response, error) {
				input := r.URL.Query().Get(optionEnsureDBsExist)
				expected := []string{"foo", "bar", "baz"}
				result := []string{}
				err := json.Unmarshal([]byte(input), &result)
				if err != nil {
					t.Fatalf("Failed to parse `%s`: %s\n", input, err)
				}
				if d := testy.DiffInterface(expected, result); d != nil {
					t.Errorf("Unexpected db list:\n%s", d)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					ProtoMajor: 1,
					ProtoMinor: 1,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(`{"state":"cluster_finished"}`)),
				}, nil
			}),
			options: map[string]interface{}{
				optionEnsureDBsExist: `["foo","bar","baz"]`,
			},
			expected: "cluster_finished",
		}
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.ClusterStatus(context.Background(), test.options)
		testy.StatusErrorRE(t, test.err, test.status, err)
		if result != test.expected {
			t.Errorf("Unexpected result:\nExpected: %s\n  Actual: %s\n", test.expected, result)
		}
	})
}

func TestClusterSetup(t *testing.T) {
	type tst struct {
		client *client
		action interface{}
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("network error", tst{
		client: newTestClient(nil, errors.New("network error")),
		status: http.StatusBadGateway,
		err:    `Post "?http://example.com/_cluster_setup"?: network error`,
	})
	tests.Add("invalid action", tst{
		client: newTestClient(nil, nil),
		action: func() {},
		status: http.StatusBadRequest,
		err:    `Post "?http://example.com/_cluster_setup"?: json: unsupported type: func()`,
	})
	tests.Add("success", func(t *testing.T) interface{} {
		return tst{
			client: newCustomClient(func(r *http.Request) (*http.Response, error) {
				expected := map[string]interface{}{
					"action": "finish_cluster",
				}
				result := map[string]interface{}{}
				if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
					t.Fatal(err)
				}
				if d := testy.DiffInterface(expected, result); d != nil {
					t.Errorf("Unexpected request body:\n%s\n", d)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					ProtoMajor: 1,
					ProtoMinor: 1,
					Header: http.Header{
						"Content-Type": []string{"application/json"},
					},
					Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
				}, nil
			}),
			action: map[string]interface{}{
				"action": "finish_cluster",
			},
		}
	})
	tests.Add("already finished", tst{
		client: newTestClient(&http.Response{
			StatusCode:    http.StatusBadRequest,
			ProtoMajor:    1,
			ProtoMinor:    1,
			ContentLength: 63,
			Header: http.Header{
				"Content-Type": []string{"application/json"},
			},
			Body: io.NopCloser(strings.NewReader(`{"error":"bad_request","reason":"Cluster is already finished"}`)),
		}, nil),
		action: map[string]interface{}{
			"action": "finish_cluster",
		},
		status: http.StatusBadRequest,
		err:    "Bad Request: Cluster is already finished",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		err := test.client.ClusterSetup(context.Background(), test.action)
		testy.StatusErrorRE(t, test.err, test.status, err)
	})
}

func TestMembership(t *testing.T) {
	type tt struct {
		client *client
		want   *driver.ClusterMembership
		status int
		err    string
	}

	tests := testy.NewTable()
	tests.Add("network error", tt{
		client: newTestClient(nil, errors.New("network error")),
		status: http.StatusBadGateway,
		err:    `Get "?http://example.com/_membership"?: network error`,
	})
	tests.Add("success 2.3.1", func(t *testing.T) interface{} {
		return tt{
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Cache-Control":  []string{"must-revalidate"},
					"Content-Length": []string{"382"},
					"Content-Type":   []string{"application/json"},
					"Date":           []string{"Fri, 10 Jul 2020 13:12:10 GMT"},
					"Server":         []string{"CouchDB/2.3.1 (Erlang OTP/19)"},
				},
				Body: io.NopCloser(strings.NewReader(`{"all_nodes":["couchdb@b2c-couchdb-0.b2c-couchdb.b2c.svc.cluster.local","couchdb@b2c-couchdb-1.b2c-couchdb.b2c.svc.cluster.local","couchdb@b2c-couchdb-2.b2c-couchdb.b2c.svc.cluster.local"],"cluster_nodes":["couchdb@b2c-couchdb-0.b2c-couchdb.b2c.svc.cluster.local","couchdb@b2c-couchdb-1.b2c-couchdb.b2c.svc.cluster.local","couchdb@b2c-couchdb-2.b2c-couchdb.b2c.svc.cluster.local"]}
				`)),
			}, nil),
			want: &driver.ClusterMembership{
				AllNodes: []string{
					"couchdb@b2c-couchdb-0.b2c-couchdb.b2c.svc.cluster.local",
					"couchdb@b2c-couchdb-1.b2c-couchdb.b2c.svc.cluster.local",
					"couchdb@b2c-couchdb-2.b2c-couchdb.b2c.svc.cluster.local",
				},
				ClusterNodes: []string{
					"couchdb@b2c-couchdb-0.b2c-couchdb.b2c.svc.cluster.local",
					"couchdb@b2c-couchdb-1.b2c-couchdb.b2c.svc.cluster.local",
					"couchdb@b2c-couchdb-2.b2c-couchdb.b2c.svc.cluster.local",
				},
			},
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		got, err := tt.client.Membership(context.Background())
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(tt.want, got); d != nil {
			t.Error(d)
		}
	})
}
