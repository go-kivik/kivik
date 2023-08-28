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
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
)

func TestAllDBs(t *testing.T) {
	tests := []struct {
		name     string
		client   *client
		options  map[string]interface{}
		expected []string
		status   int
		err      string
	}{
		{
			name:   "network error",
			client: newTestClient(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/_all_dbs"?: net error`,
		},
		{
			name: "2.0.0",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Server":              {"CouchDB/2.0.0 (Erlang OTP/17)"},
					"Date":                {"Fri, 27 Oct 2017 15:15:07 GMT"},
					"Content-Type":        {"application/json"},
					"ETag":                {`"33UVNAZU752CYNGBBTMWQFP7U"`},
					"Transfer-Encoding":   {"chunked"},
					"X-Couch-Request-ID":  {"ab5cd97c3e"},
					"X-CouchDB-Body-Time": {"0"},
				},
				Body: Body(`["_global_changes","_replicator","_users"]`),
			}, nil),
			expected: []string{"_global_changes", "_replicator", "_users"},
		},
		{
			name:    "bad options",
			options: map[string]interface{}{"foo": func() {}},
			status:  http.StatusBadRequest,
			err:     `kivik: invalid type func\(\) for options`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.AllDBs(context.Background(), test.options)
			testy.StatusErrorRE(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestDBExists(t *testing.T) {
	tests := []struct {
		name   string
		client *client
		dbName string
		exists bool
		status int
		err    string
	}{
		{
			name:   "no db specified",
			status: http.StatusBadRequest,
			err:    "kivik: dbName required",
		},
		{
			name:   "network error",
			dbName: "foo",
			client: newTestClient(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Head "?http://example.com/foo"?: net error`,
		},
		{
			name:   "not found, 1.6.1",
			dbName: "foox",
			client: newTestClient(&http.Response{
				StatusCode: 404,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":           {"Fri, 27 Oct 2017 15:09:19 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"44"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: Body(""),
			}, nil),
			exists: false,
		},
		{
			name:   "exists, 1.6.1",
			dbName: "foo",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":           {"Fri, 27 Oct 2017 15:09:19 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"229"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: Body(""),
			}, nil),
			exists: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exists, err := test.client.DBExists(context.Background(), test.dbName, nil)
			testy.StatusErrorRE(t, test.err, test.status, err)
			if exists != test.exists {
				t.Errorf("Unexpected result: %t", exists)
			}
		})
	}
}

func TestCreateDB(t *testing.T) {
	tests := []struct {
		name    string
		dbName  string
		options map[string]interface{}
		client  *client
		status  int
		err     string
	}{
		{
			name:   "missing dbname",
			status: http.StatusBadRequest,
			err:    "kivik: dbName required",
		},
		{
			name:   "network error",
			dbName: "foo",
			client: newTestClient(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Put "?http://example.com/foo"?: net error`,
		},
		{
			name:   "conflict 1.6.1",
			dbName: "foo",
			client: newTestClient(&http.Response{
				StatusCode: 412,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":           {"Fri, 27 Oct 2017 15:23:57 GMT"},
					"Content-Type":   {"application/json"},
					"Content-Length": {"94"},
					"Cache-Control":  {"must-revalidate"},
				},
				ContentLength: 94,
				Body:          Body(`{"error":"file_exists","reason":"The database could not be created, the file already exists."}`),
			}, nil),
			status: http.StatusPreconditionFailed,
			err:    "Precondition Failed: The database could not be created, the file already exists.",
		},
		{
			name:    "bad options",
			dbName:  "foo",
			options: map[string]interface{}{"foo": func() {}},
			status:  http.StatusBadRequest,
			err:     `^kivik: invalid type func\(\) for options$`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.client.CreateDB(context.Background(), test.dbName, test.options)
			testy.StatusErrorRE(t, test.err, test.status, err)
		})
	}
}

func TestDestroyDB(t *testing.T) {
	tests := []struct {
		name   string
		client *client
		dbName string
		status int
		err    string
	}{
		{
			name:   "no db name",
			status: http.StatusBadRequest,
			err:    "kivik: dbName required",
		},
		{
			name:   "network error",
			dbName: "foo",
			client: newTestClient(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `(Delete "?http://example.com/foo"?: )?net error`,
		},
		{
			name:   "1.6.1",
			dbName: "foo",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":           {"Fri, 27 Oct 2017 17:12:45 GMT"},
					"Content-Type":   {"application/json"},
					"Content-Length": {"12"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: Body(`{"ok":true}`),
			}, nil),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.client.DestroyDB(context.Background(), test.dbName, nil)
			testy.StatusErrorRE(t, test.err, test.status, err)
		})
	}
}

func TestDBUpdates(t *testing.T) {
	tests := []struct {
		name   string
		client *client
		status int
		err    string
	}{
		{
			name:   "network error",
			client: newTestClient(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/_db_updates\?feed=continuous&since=now"?: net error`,
		},
		{
			name: "error response",
			client: newTestClient(&http.Response{
				StatusCode: 400,
				Body:       Body(""),
			}, nil),
			status: http.StatusBadRequest,
			err:    "Bad Request",
		},
		{
			name: "Success 1.6.1",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Transfer-Encoding": {"chunked"},
					"Server":            {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":              {"Fri, 27 Oct 2017 19:55:43 GMT"},
					"Content-Type":      {"application/json"},
					"Cache-Control":     {"must-revalidate"},
				},
				Body: Body(""),
			}, nil),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.DBUpdates(context.TODO(), nil)
			testy.StatusErrorRE(t, test.err, test.status, err)
			if _, ok := result.(*couchUpdates); !ok {
				t.Errorf("Unexpected type returned: %t", result)
			}
		})
	}
}

func TestUpdatesNext(t *testing.T) {
	tests := []struct {
		name     string
		updates  *couchUpdates
		status   int
		err      string
		expected *driver.DBUpdate
	}{
		{
			name:     "consumed feed",
			updates:  newUpdates(context.TODO(), Body("")),
			expected: &driver.DBUpdate{},
			status:   http.StatusInternalServerError,
			err:      "EOF",
		},
		{
			name:    "read feed",
			updates: newUpdates(context.TODO(), Body(`{"db_name":"mailbox","type":"created","seq":"1-g1AAAAFReJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuDOZExFyjAnmJhkWaeaIquGIf2JAUgmWQPMiGRAZcaB5CaePxqEkBq6vGqyWMBkgwNQAqobD4h"},`)),
			expected: &driver.DBUpdate{
				DBName: "mailbox",
				Type:   "created",
				Seq:    "1-g1AAAAFReJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuDOZExFyjAnmJhkWaeaIquGIf2JAUgmWQPMiGRAZcaB5CaePxqEkBq6vGqyWMBkgwNQAqobD4h",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := new(driver.DBUpdate)
			err := test.updates.Next(result)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestUpdatesClose(t *testing.T) {
	body := &closeTracker{ReadCloser: Body("")}
	u := newUpdates(context.TODO(), body)
	if err := u.Close(); err != nil {
		t.Fatal(err)
	}
	if !body.closed {
		t.Errorf("Failed to close")
	}
}

func TestPing(t *testing.T) {
	type pingTest struct {
		name     string
		client   *client
		expected bool
		status   int
		err      string
	}

	tests := []pingTest{
		{
			name: "Couch 1.6",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusBadRequest,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header: http.Header{
					"Server": []string{"CouchDB/1.6.1 (Erlang OTP/17)"},
				},
			}, nil),
			expected: true,
		},
		{
			name: "Couch 2.x offline",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusNotFound,
				ProtoMajor: 1,
				ProtoMinor: 1,
			}, nil),
			expected: false,
		},
		{
			name: "Couch 2.x online",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				ProtoMajor: 1,
				ProtoMinor: 1,
			}, nil),
			expected: true,
		},
		{
			name:     "network error",
			client:   newTestClient(nil, errors.New("network error")),
			expected: false,
			status:   http.StatusBadGateway,
			err:      `Head "?http://example.com/_up"?: network error`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.Ping(context.Background())
			testy.StatusErrorRE(t, test.err, test.status, err)
			if result != test.expected {
				t.Errorf("Unexpected result: %t", result)
			}
		})
	}
}
