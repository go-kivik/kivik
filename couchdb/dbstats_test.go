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
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
)

func TestStats(t *testing.T) {
	tests := []struct {
		name     string
		db       *db
		expected *driver.DBStats
		status   int
		err      string
	}{
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/testdb"?: net error`,
		},
		{
			name: "read error",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Body: &mockReadCloser{
					ReadFunc: func(_ []byte) (int, error) {
						return 0, errors.New("read error")
					},
					CloseFunc: func() error { return nil },
				},
			}, nil),
			status: http.StatusBadGateway,
			err:    "read error",
		},
		{
			name: "invalid JSON response",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`invalid json`)),
			}, nil),
			status: http.StatusBadGateway,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "error response",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil),
			status: http.StatusBadRequest,
			err:    "Bad Request",
		},
		{
			name: "1.6.1",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":           {"Thu, 26 Oct 2017 12:58:14 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"235"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: io.NopCloser(strings.NewReader(`{"db_name":"_users","doc_count":3,"doc_del_count":14,"update_seq":31,"purge_seq":0,"compact_running":false,"disk_size":127080,"data_size":6028,"instance_start_time":"1509022681259533","disk_format_version":6,"committed_update_seq":31}`)),
			}, nil),
			expected: &driver.DBStats{
				Name:         "_users",
				DocCount:     3,
				DeletedCount: 14,
				UpdateSeq:    "31",
				DiskSize:     127080,
				ActiveSize:   6028,
				RawResponse:  []byte(`{"db_name":"_users","doc_count":3,"doc_del_count":14,"update_seq":31,"purge_seq":0,"compact_running":false,"disk_size":127080,"data_size":6028,"instance_start_time":"1509022681259533","disk_format_version":6,"committed_update_seq":31}`),
			},
		},
		{
			name: "2.0.0",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Server":              {"CouchDB/2.0.0 (Erlang OTP/17)"},
					"Date":                {"Thu, 26 Oct 2017 13:01:13 GMT"},
					"Content-Type":        {"application/json"},
					"Content-Length":      {"429"},
					"Cache-Control":       {"must-revalidate"},
					"X-Couch-Request-ID":  {"2486f27546"},
					"X-CouchDB-Body-Time": {"0"},
				},
				Body: io.NopCloser(strings.NewReader(`{"db_name":"_users","update_seq":"13-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWQPVsOCS40DSE08WA0rLjUJIDX1eO3KYwGSDA1ACqhsPiF1CyDq9mclMuFVdwCi7j4hdQ8g6kDuywIAkRBjAw","sizes":{"file":87323,"external":2495,"active":6082},"purge_seq":0,"other":{"data_size":2495},"doc_del_count":6,"doc_count":1,"disk_size":87323,"disk_format_version":6,"data_size":6082,"compact_running":false,"instance_start_time":"0"}`)),
			}, nil),
			expected: &driver.DBStats{
				Name:         "_users",
				DocCount:     1,
				DeletedCount: 6,
				UpdateSeq:    "13-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWQPVsOCS40DSE08WA0rLjUJIDX1eO3KYwGSDA1ACqhsPiF1CyDq9mclMuFVdwCi7j4hdQ8g6kDuywIAkRBjAw",
				DiskSize:     87323,
				ActiveSize:   6082,
				ExternalSize: 2495,
				RawResponse:  []byte(`{"db_name":"_users","update_seq":"13-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWQPVsOCS40DSE08WA0rLjUJIDX1eO3KYwGSDA1ACqhsPiF1CyDq9mclMuFVdwCi7j4hdQ8g6kDuywIAkRBjAw","sizes":{"file":87323,"external":2495,"active":6082},"purge_seq":0,"other":{"data_size":2495},"doc_del_count":6,"doc_count":1,"disk_size":87323,"disk_format_version":6,"data_size":6082,"compact_running":false,"instance_start_time":"0"}`),
			},
		},
		{
			name: "2.1.1",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Server":              {"CouchDB/2.0.0 (Erlang OTP/17)"},
					"Date":                {"Thu, 26 Oct 2017 13:01:13 GMT"},
					"Content-Type":        {"application/json"},
					"Content-Length":      {"429"},
					"Cache-Control":       {"must-revalidate"},
					"X-Couch-Request-ID":  {"2486f27546"},
					"X-CouchDB-Body-Time": {"0"},
				},
				Body: io.NopCloser(strings.NewReader(`{"db_name":"_users","update_seq":"13-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWQPVsOCS40DSE08WA0rLjUJIDX1eO3KYwGSDA1ACqhsPiF1CyDq9mclMuFVdwCi7j4hdQ8g6kDuywIAkRBjAw","sizes":{"file":87323,"external":2495,"active":6082},"purge_seq":0,"other":{"data_size":2495},"doc_del_count":6,"doc_count":1,"disk_size":87323,"disk_format_version":6,"data_size":6082,"compact_running":false,"instance_start_time":"0","cluster":{"n":1,"q":2,"r":3,"w":4}}`)),
			}, nil),
			expected: &driver.DBStats{
				Name:         "_users",
				DocCount:     1,
				DeletedCount: 6,
				UpdateSeq:    "13-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWQPVsOCS40DSE08WA0rLjUJIDX1eO3KYwGSDA1ACqhsPiF1CyDq9mclMuFVdwCi7j4hdQ8g6kDuywIAkRBjAw",
				DiskSize:     87323,
				ActiveSize:   6082,
				ExternalSize: 2495,
				Cluster: &driver.ClusterStats{
					Replicas:    1,
					Shards:      2,
					ReadQuorum:  3,
					WriteQuorum: 4,
				},
				RawResponse: []byte(`{"db_name":"_users","update_seq":"13-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWQPVsOCS40DSE08WA0rLjUJIDX1eO3KYwGSDA1ACqhsPiF1CyDq9mclMuFVdwCi7j4hdQ8g6kDuywIAkRBjAw","sizes":{"file":87323,"external":2495,"active":6082},"purge_seq":0,"other":{"data_size":2495},"doc_del_count":6,"doc_count":1,"disk_size":87323,"disk_format_version":6,"data_size":6082,"compact_running":false,"instance_start_time":"0","cluster":{"n":1,"q":2,"r":3,"w":4}}`),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Stats(context.Background())
			testy.StatusErrorRE(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestDbsStats(t *testing.T) {
	tests := []struct {
		name     string
		client   *client
		dbnames  []string
		expected interface{}
		status   int
		err      string
	}{
		{
			name:    "network error",
			client:  newTestClient(nil, errors.New("net error")),
			dbnames: []string{"foo", "bar"},
			status:  http.StatusBadGateway,
			err:     `Post "?http://example.com/_dbs_info"?: net error`,
		},
		{
			name: "read error",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Body: &mockReadCloser{
					ReadFunc: func(_ []byte) (int, error) {
						return 0, errors.New("read error")
					},
					CloseFunc: func() error { return nil },
				},
			}, nil),
			status: http.StatusBadGateway,
			err:    "read error",
		},
		{
			name: "invalid JSON response",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`invalid json`)),
			}, nil),
			status: http.StatusBadGateway,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "error response",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil),
			status: http.StatusBadRequest,
			err:    "Bad Request",
		},
		{
			name: "2.1.2",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusNotFound,
				Header: http.Header{
					"Server":              {"CouchDB/2.1.2 (Erlang OTP/17)"},
					"Date":                {"Sat, 01 Sep 2018 15:42:53 GMT"},
					"Content-Type":        {"application/json"},
					"Content-Length":      {"58"},
					"Cache-Control":       {"must-revalidate"},
					"X-Couch-Request-ID":  {"e1264663f9"},
					"X-CouchDB-Body-Time": {"0"},
				},
				Body: io.NopCloser(strings.NewReader(`{"error":"not_found","reason":"Database does not exist."}`)),
			}, nil),
			dbnames: []string{"foo", "bar"},
			err:     "Not Found",
			status:  http.StatusNotFound,
		},
		{
			name: "2.2.0",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Server":              {"CouchDB/2.2.0 (Erlang OTP/19)"},
					"Date":                {"Sat, 01 Sep 2018 15:50:56 GMT"},
					"Content-Type":        {"application/json"},
					"Transfer-Encoding":   {"chunked"},
					"Cache-Control":       {"must-revalidate"},
					"X-Couch-Request-ID":  {"1bf258cfbe"},
					"X-CouchDB-Body-Time": {"0"},
				},
				Body: io.NopCloser(strings.NewReader(`[{"key":"foo","error":"not_found"},{"key":"bar","error":"not_found"},{"key":"_users","info":{"db_name":"_users","update_seq":"1-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWSPX40DSE08WA0jLjUJIDX1eM3JYwGSDA1ACqhsPiF1CyDq9hNSdwCi7j4hdQ8g6kDuywIAiVhi9w","sizes":{"file":24423,"external":5361,"active":2316},"purge_seq":0,"other":{"data_size":5361},"doc_del_count":0,"doc_count":1,"disk_size":24423,"disk_format_version":6,"data_size":2316,"compact_running":false,"cluster":{"q":8,"n":1,"w":1,"r":1},"instance_start_time":"0"}}]
`)),
			}, nil),
			expected: []*driver.DBStats{
				nil,
				nil,
				{
					Name:         "_users",
					DocCount:     1,
					UpdateSeq:    "1-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWSPX40DSE08WA0jLjUJIDX1eM3JYwGSDA1ACqhsPiF1CyDq9hNSdwCi7j4hdQ8g6kDuywIAiVhi9w",
					DiskSize:     24423,
					ActiveSize:   2316,
					ExternalSize: 5361,
					Cluster: &driver.ClusterStats{
						Replicas:    1,
						Shards:      8,
						ReadQuorum:  1,
						WriteQuorum: 1,
					},
					RawResponse: []byte(`{"db_name":"_users","update_seq":"1-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuRAYeCJAUgmWSPX40DSE08WA0jLjUJIDX1eM3JYwGSDA1ACqhsPiF1CyDq9hNSdwCi7j4hdQ8g6kDuywIAiVhi9w","sizes":{"file":24423,"external":5361,"active":2316},"purge_seq":0,"other":{"data_size":5361},"doc_del_count":0,"doc_count":1,"disk_size":24423,"disk_format_version":6,"data_size":2316,"compact_running":false,"cluster":{"q":8,"n":1,"w":1,"r":1},"instance_start_time":"0"}`),
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.DBsStats(context.Background(), test.dbnames)
			testy.StatusErrorRE(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestPartitionStats(t *testing.T) {
	type tt struct {
		db     *db
		name   string
		status int
		err    string
	}

	tests := testy.NewTable()
	tests.Add("network error", tt{
		db:     newTestDB(nil, errors.New("net error")),
		name:   "partXX",
		status: http.StatusBadGateway,
		err:    `Get "?http://example.com/testdb/_partition/partXX"?: net error`,
	})
	tests.Add("read error", tt{
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Body: &mockReadCloser{
				ReadFunc: func(_ []byte) (int, error) {
					return 0, errors.New("read error")
				},
				CloseFunc: func() error { return nil },
			},
		}, nil),
		status: http.StatusBadGateway,
		err:    "read error",
	})
	tests.Add("invalid JSON response", tt{
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`invalid json`)),
		}, nil),
		status: http.StatusBadGateway,
		err:    "invalid character 'i' looking for beginning of value",
	})
	tests.Add("error response", tt{
		db: newTestDB(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil),
		status: http.StatusBadRequest,
		err:    "Bad Request",
	})
	tests.Add("3.0.0-pre", tt{
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Server":              {"CouchDB/2.3.0-a1e11cea9 (Erlang OTP/21)"},
				"Date":                {"Thu, 24 Jan 2019 17:19:59 GMT"},
				"Content-Type":        {"application/json"},
				"Content-Length":      {"119"},
				"Cache-Control":       {"must-revalidate"},
				"X-Couch-Request-ID":  {"2486f27546"},
				"X-CouchDB-Body-Time": {"0"},
			},
			Body: io.NopCloser(strings.NewReader(`{"db_name":"my_new_db","doc_count":1,"doc_del_count":0,"partition":"sensor-260","sizes":{"active":244,"external":347}}
`)),
		}, nil),
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		result, err := tt.db.PartitionStats(context.Background(), tt.name)
		testy.StatusErrorRE(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}
