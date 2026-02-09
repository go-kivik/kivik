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
	"fmt"
	"io"
	"net/http"
	"net/url"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestAllDBs(t *testing.T) {
	type test struct {
		client   *client
		options  kivik.Option
		expected []string
		status   int
		err      string
	}
	tests := testy.NewTable()
	tests.Add("network error", test{
		client: newTestClient(nil, errors.New("net error")),
		status: http.StatusBadGateway,
		err:    `Get "?http://example.com/_all_dbs"?: net error`,
	})
	tests.Add("2.0.0", test{
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
	})
	tests.Add("with param", test{
		client:  newTestClient(nil, errors.New("expected")),
		options: kivik.Param("startkey", "bar"),
		status:  http.StatusBadGateway,
		err:     `Get "?http://example.com/_all_dbs\?startkey=bar"?: expected`,
	})

	tests.Run(t, func(t *testing.T, test test) {
		opts := test.options
		if opts == nil {
			opts = mock.NilOption
		}
		result, err := test.client.AllDBs(context.Background(), opts)
		if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
			t.Error(d)
		}
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
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
		{
			name:   "slashes",
			dbName: "foo/bar",
			client: newCustomClient(func(req *http.Request) (*http.Response, error) {
				if err := consume(req.Body); err != nil {
					return nil, err
				}
				expected := "/" + url.PathEscape("foo/bar")
				actual := req.URL.RawPath
				if actual != expected {
					return nil, fmt.Errorf("expected path %s, got %s", expected, actual)
				}
				response := &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
						"Date":           {"Fri, 27 Oct 2017 15:09:19 GMT"},
						"Content-Type":   {"text/plain; charset=utf-8"},
						"Content-Length": {"229"},
						"Cache-Control":  {"must-revalidate"},
					},
					Body: Body(""),
				}
				response.Request = req
				return response, nil
			}),
			exists: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			exists, err := test.client.DBExists(context.Background(), test.dbName, nil)
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
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
		options kivik.Option
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := test.options
			if opts == nil {
				opts = mock.NilOption
			}
			err := test.client.CreateDB(context.Background(), test.dbName, opts)
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
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
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestDBUpdates(t *testing.T) {
	tests := []struct {
		name       string
		client     *client
		options    driver.Options
		want       []driver.DBUpdate
		wantStatus int
		wantErr    string
	}{
		{
			name:       "network error",
			client:     newTestClient(nil, errors.New("net error")),
			wantStatus: http.StatusBadGateway,
			wantErr:    `Get "?http://example.com/_db_updates\?feed=continuous&since=now"?: net error`,
		},
		{
			name: "CouchDB defaults, network error",
			options: kivik.Params(map[string]any{
				"feed":  "",
				"since": "",
			}),
			client:     newTestClient(nil, errors.New("net error")),
			wantStatus: http.StatusBadGateway,
			wantErr:    `Get "?http://example.com/_db_updates"?: net error`,
		},
		{
			name: "error response",
			client: newTestClient(&http.Response{
				StatusCode: 400,
				Body:       Body(""),
			}, nil),
			wantStatus: http.StatusBadRequest,
			wantErr:    "Bad Request",
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
				Body: Body(`{"db_name":"mailbox","type":"created","seq":"1-g1AAAAFR"}
				{"db_name":"mailbox","type":"deleted","seq":"2-g1AAAAFR"}`),
			}, nil),
			want: []driver.DBUpdate{
				{DBName: "mailbox", Type: "created", Seq: "1-g1AAAAFR"},
				{DBName: "mailbox", Type: "deleted", Seq: "2-g1AAAAFR"},
			},
		},
		{
			name: "non-JSON response",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Transfer-Encoding": {"chunked"},
					"Server":            {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":              {"Fri, 27 Oct 2017 19:55:43 GMT"},
					"Content-Type":      {"application/json"},
					"Cache-Control":     {"must-revalidate"},
				},
				Body: Body(`invalid json`),
			}, nil),
			wantStatus: http.StatusBadGateway,
			wantErr:    `invalid character 'i' looking for beginning of value`,
		},
		{
			name: "wrong opening JSON token",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Transfer-Encoding": {"chunked"},
					"Server":            {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":              {"Fri, 27 Oct 2017 19:55:43 GMT"},
					"Content-Type":      {"application/json"},
					"Cache-Control":     {"must-revalidate"},
				},
				Body: Body(`[]`),
			}, nil),
			wantStatus: http.StatusBadGateway,
			wantErr:    "expected `{`",
		},
		{
			name: "wrong second JSON token type",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Transfer-Encoding": {"chunked"},
					"Server":            {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":              {"Fri, 27 Oct 2017 19:55:43 GMT"},
					"Content-Type":      {"application/json"},
					"Cache-Control":     {"must-revalidate"},
				},
				Body: Body(`{"foo":"bar"}`),
			}, nil),
			wantStatus: http.StatusBadGateway,
			wantErr:    "expected `db_name` or `results`",
		},
		{
			name: "CouchDB defaults",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Transfer-Encoding": {"chunked"},
					"Server":            {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":              {"Fri, 27 Oct 2017 19:55:43 GMT"},
					"Content-Type":      {"application/json"},
					"Cache-Control":     {"must-revalidate"},
				},
				Body: Body(`{
					"results":[
						{"db_name":"mailbox","type":"created","seq":"1-g1AAAAFR"},
						{"db_name":"mailbox","type":"deleted","seq":"2-g1AAAAFR"}
					],
					"last_seq": "2-g1AAAAFR"
				}`),
			}, nil),
			options: kivik.Params(map[string]any{
				"feed":  "",
				"since": "",
			}),
			want: []driver.DBUpdate{
				{DBName: "mailbox", Type: "created", Seq: "1-g1AAAAFR"},
				{DBName: "mailbox", Type: "deleted", Seq: "2-g1AAAAFR"},
			},
		},
		{
			name: "eventsource",
			options: kivik.Params(map[string]any{
				"feed":  "eventsource",
				"since": "",
			}),
			wantStatus: http.StatusBadRequest,
			wantErr:    "eventsource feed type not supported",
		},
		{
			// Based on CI test failures, presumably from a race condition that
			// causes the query to happen before any database is created.
			name: "no databases",
			client: newTestClient(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Content-Type": {"application/json"},
				},
				Body: Body(`{"last_seq":"38-g1AAAACLeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5UTgXKMBuZmFmYWFgjq4Yh_Y8FiDJ0ACk_qOYYpyanGiQYoquJwsAM_UqgA"}`),
			}, nil),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := tt.options
			if opts == nil {
				opts = mock.NilOption
			}
			result, err := tt.client.DBUpdates(context.TODO(), opts)
			if d := internal.StatusErrorDiffRE(tt.wantErr, tt.wantStatus, err); d != "" {
				t.Error(d)
			}
			if err != nil {
				return
			}

			var got []driver.DBUpdate
			for {
				var update driver.DBUpdate
				err := result.Next(&update)
				if err == io.EOF {
					break
				}
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, update)
			}
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf("Unexpected result:\n%s\n", d)
			}
		})
	}
}

func Test_updatesForFeedType(t *testing.T) {
	t.Parallel()

	type test struct {
		ft         feedType
		wantStatus int
		wantErr    string
	}

	tests := testy.NewTable()
	tests.Add("unknown feed type", test{
		ft:         feedType(99),
		wantStatus: http.StatusBadGateway,
		wantErr:    `unknown feed type`,
	})

	tests.Run(t, func(t *testing.T, tt test) {
		_, err := updatesForFeedType(t.Context(), Body(""), tt.ft)
		if d := internal.StatusErrorDiffRE(tt.wantErr, tt.wantStatus, err); d != "" {
			t.Error(d)
		}
	})
}

func newTestUpdates(t *testing.T, body io.ReadCloser) driver.DBUpdates {
	t.Helper()
	u, err := newUpdates(context.Background(), body)
	if err != nil {
		t.Fatal(err)
	}
	return u
}

func TestUpdatesNext(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		updates  driver.DBUpdates
		status   int
		err      string
		expected *driver.DBUpdate
	}{
		{
			name:     "consumed feed",
			updates:  newContinuousUpdates(context.TODO(), Body("")),
			expected: &driver.DBUpdate{},
			status:   http.StatusInternalServerError,
			err:      "EOF",
		},
		{
			name:    "read feed",
			updates: newTestUpdates(t, Body(`{"db_name":"mailbox","type":"created","seq":"1-g1AAAAFReJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjMlMiTJ____PyuDOZExFyjAnmJhkWaeaIquGIf2JAUgmWQPMiGRAZcaB5CaePxqEkBq6vGqyWMBkgwNQAqobD4h"},`)),
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
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestUpdatesClose(t *testing.T) {
	t.Parallel()
	body := &closeTracker{ReadCloser: Body("")}
	u := newContinuousUpdates(context.TODO(), body)
	if err := u.Close(); err != nil {
		t.Fatal(err)
	}
	if !body.closed {
		t.Errorf("Failed to close")
	}
}

func TestUpdatesLastSeq(t *testing.T) {
	t.Parallel()

	client := newTestClient(&http.Response{
		StatusCode: 200,
		Header: http.Header{
			"Transfer-Encoding": {"chunked"},
			"Server":            {"CouchDB/1.6.1 (Erlang OTP/17)"},
			"Date":              {"Fri, 27 Oct 2017 19:55:43 GMT"},
			"Content-Type":      {"application/json"},
			"Cache-Control":     {"must-revalidate"},
		},
		Body: Body(`{"results":[],"last_seq":"99-asdf"}`),
	}, nil)

	updates, err := client.DBUpdates(context.TODO(), mock.NilOption)
	if err != nil {
		t.Fatal(err)
	}
	for {
		err := updates.Next(&driver.DBUpdate{})
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

	}
	want := "99-asdf"
	got, err := updates.(driver.LastSeqer).LastSeq()
	if err != nil {
		t.Fatal(err)
	}
	if got != want {
		t.Errorf("Unexpected last_seq: %s", got)
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
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if result != test.expected {
				t.Errorf("Unexpected result: %t", result)
			}
		})
	}
}
