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
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
)

func TestExplain(t *testing.T) {
	tests := []struct {
		name     string
		db       *db
		query    interface{}
		opts     map[string]interface{}
		expected *driver.QueryPlan
		status   int
		err      string
	}{
		{
			name:   "invalid query",
			db:     newTestDB(nil, nil),
			query:  make(chan int),
			status: http.StatusBadRequest,
			err:    `Post "?http://example.com/testdb/_explain"?: json: unsupported type: chan int`,
		},
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_explain"?: net error`,
		},
		{
			name: "error response",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusNotFound,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil),
			status: http.StatusNotFound,
			err:    "Not Found",
		},
		{
			name: "success",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"dbname":"foo"}`)),
			}, nil),
			expected: &driver.QueryPlan{DBName: "foo"},
		},
		{
			name: "raw query",
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				defer req.Body.Close() // nolint: errcheck
				var result interface{}
				if err := json.NewDecoder(req.Body).Decode(&result); err != nil {
					return nil, fmt.Errorf("decode error: %s", err)
				}
				expected := map[string]interface{}{"_id": "foo"}
				if d := testy.DiffInterface(expected, result); d != nil {
					return nil, fmt.Errorf("unexpected result:\n%s", d)
				}
				return nil, errors.New("success")
			}),
			query:  []byte(`{"_id":"foo"}`),
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_explain"?: success`,
		},
		{
			name: "partitioned request",
			db:   newTestDB(nil, errors.New("expected")),
			opts: map[string]interface{}{
				OptionPartition: "x1",
			},
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_partition/x1/_explain"?: expected`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Explain(context.Background(), test.query, test.opts)
			testy.StatusErrorRE(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestUnmarshalQueryPlan(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected *queryPlan
		err      string
	}{
		{
			name:  "non-array",
			input: `{"fields":{}}`,
			err:   "json: cannot unmarshal object into Go",
		},
		{
			name:     "all_fields",
			input:    `{"fields":"all_fields","dbname":"foo"}`,
			expected: &queryPlan{DBName: "foo"},
		},
		{
			name:     "simple field list",
			input:    `{"fields":["foo","bar"],"dbname":"foo"}`,
			expected: &queryPlan{Fields: []interface{}{"foo", "bar"}, DBName: "foo"},
		},
		{
			name:  "complex field list",
			input: `{"dbname":"foo", "fields":[{"foo":"asc"},{"bar":"desc"}]}`,
			expected: &queryPlan{
				DBName: "foo",
				Fields: []interface{}{
					map[string]interface{}{"foo": "asc"},
					map[string]interface{}{"bar": "desc"},
				},
			},
		},
		{
			name:  "invalid bare string",
			input: `{"fields":"not_all_fields"}`,
			err:   "json: cannot unmarshal string into Go",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := new(queryPlan)
			err := json.Unmarshal([]byte(test.input), &result)
			testy.ErrorRE(t, test.err, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestCreateIndex(t *testing.T) {
	tests := []struct {
		name            string
		ddoc, indexName string
		index           interface{}
		options         map[string]interface{}
		db              *db
		status          int
		err             string
	}{
		{
			name:   "invalid JSON index",
			db:     newTestDB(nil, nil),
			index:  `invalid json`,
			status: http.StatusBadRequest,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name:   "invalid raw index",
			db:     newTestDB(nil, nil),
			index:  map[string]interface{}{"foo": make(chan int)},
			status: http.StatusBadRequest,
			err:    `Post "?http://example.com/testdb/_index"?: json: unsupported type: chan int`,
		},
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_index"?: net error`,
		},
		{
			name: "success 2.1.0",
			db: newTestDB(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"X-CouchDB-Body-Time": {"0"},
					"X-Couch-Request-ID":  {"8e4aef0c2f"},
					"Server":              {"CouchDB/2.1.0 (Erlang OTP/17)"},
					"Date":                {"Fri, 27 Oct 2017 18:14:38 GMT"},
					"Content-Type":        {"application/json"},
					"Content-Length":      {"126"},
					"Cache-Control":       {"must-revalidate"},
				},
				Body: Body(`{"result":"created","id":"_design/a7ee061f1a2c0c6882258b2f1e148b714e79ccea","name":"a7ee061f1a2c0c6882258b2f1e148b714e79ccea"}`),
			}, nil),
		},
		{
			name: "partitioned query",
			db:   newTestDB(nil, errors.New("expected")),
			options: map[string]interface{}{
				OptionPartition: "xxy",
			},
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_partition/xxy/_index"?: expected`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.CreateIndex(context.Background(), test.ddoc, test.indexName, test.index, test.options)
			testy.StatusErrorRE(t, test.err, test.status, err)
		})
	}
}

func TestGetIndexes(t *testing.T) {
	tests := []struct {
		name     string
		options  map[string]interface{}
		db       *db
		expected []driver.Index
		status   int
		err      string
	}{
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/testdb/_index"?: net error`,
		},
		{
			name: "2.1.0",
			db: newTestDB(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"X-CouchDB-Body-Time": {"0"},
					"X-Couch-Request-ID":  {"f44881735c"},
					"Server":              {"CouchDB/2.1.0 (Erlang OTP/17)"},
					"Date":                {"Fri, 27 Oct 2017 18:23:29 GMT"},
					"Content-Type":        {"application/json"},
					"Content-Length":      {"269"},
					"Cache-Control":       {"must-revalidate"},
				},
				Body: Body(`{"total_rows":2,"indexes":[{"ddoc":null,"name":"_all_docs","type":"special","def":{"fields":[{"_id":"asc"}]}},{"ddoc":"_design/a7ee061f1a2c0c6882258b2f1e148b714e79ccea","name":"a7ee061f1a2c0c6882258b2f1e148b714e79ccea","type":"json","def":{"fields":[{"foo":"asc"}]}}]}`),
			}, nil),
			expected: []driver.Index{
				{
					Name: "_all_docs",
					Type: "special",
					Definition: map[string]interface{}{
						"fields": []interface{}{
							map[string]interface{}{"_id": "asc"},
						},
					},
				},
				{
					DesignDoc: "_design/a7ee061f1a2c0c6882258b2f1e148b714e79ccea",
					Name:      "a7ee061f1a2c0c6882258b2f1e148b714e79ccea",
					Type:      "json",
					Definition: map[string]interface{}{
						"fields": []interface{}{
							map[string]interface{}{"foo": "asc"},
						},
					},
				},
			},
		},
		{
			name: "partitioned query",
			db:   newTestDB(nil, errors.New("expected")),
			options: map[string]interface{}{
				OptionPartition: "yyz",
			},
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/testdb/_partition/yyz/_index"?: expected`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.GetIndexes(context.Background(), test.options)
			testy.StatusErrorRE(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestDeleteIndex(t *testing.T) {
	tests := []struct {
		name            string
		ddoc, indexName string
		options         map[string]interface{}
		db              *db
		status          int
		err             string
	}{
		{
			name:   "no ddoc",
			status: http.StatusBadRequest,
			db:     newTestDB(nil, nil),
			err:    "kivik: ddoc required",
		},
		{
			name:   "no index name",
			ddoc:   "foo",
			status: http.StatusBadRequest,
			db:     newTestDB(nil, nil),
			err:    "kivik: name required",
		},
		{
			name:      "network error",
			ddoc:      "foo",
			indexName: "bar",
			db:        newTestDB(nil, errors.New("net error")),
			status:    http.StatusBadGateway,
			err:       `^(Delete "?http://example.com/testdb/_index/foo/json/bar"?: )?net error`,
		},
		{
			name:      "2.1.0 success",
			ddoc:      "_design/a7ee061f1a2c0c6882258b2f1e148b714e79ccea",
			indexName: "a7ee061f1a2c0c6882258b2f1e148b714e79ccea",
			db: newTestDB(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"X-CouchDB-Body-Time": {"0"},
					"X-Couch-Request-ID":  {"6018a0a693"},
					"Server":              {"CouchDB/2.1.0 (Erlang OTP/17)"},
					"Date":                {"Fri, 27 Oct 2017 19:06:28 GMT"},
					"Content-Type":        {"application/json"},
					"Content-Length":      {"11"},
					"Cache-Control":       {"must-revalidate"},
				},
				Body: Body(`{"ok":true}`),
			}, nil),
		},
		{
			name:      "partitioned query",
			ddoc:      "_design/foo",
			indexName: "bar",
			db:        newTestDB(nil, errors.New("expected")),
			options: map[string]interface{}{
				OptionPartition: "qqz",
			},
			status: http.StatusBadGateway,
			err:    `Delete "?http://example.com/testdb/_partition/qqz/_index/_design/foo/json/bar"?: expected`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.DeleteIndex(context.Background(), test.ddoc, test.indexName, test.options)
			testy.StatusErrorRE(t, test.err, test.status, err)
		})
	}
}

func TestFind(t *testing.T) {
	tests := []struct {
		name   string
		db     *db
		query  interface{}
		opts   map[string]interface{}
		status int
		err    string
	}{
		{
			name:   "invalid query json",
			db:     newTestDB(nil, nil),
			query:  make(chan int),
			status: http.StatusBadRequest,
			err:    `Post "?http://example.com/testdb/_find"?: json: unsupported type: chan int`,
		},
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_find"?: net error`,
		},
		{
			name: "error response",
			db: newTestDB(&http.Response{
				StatusCode: 415,
				Header: http.Header{
					"Content-Type":        {"application/json"},
					"X-CouchDB-Body-Time": {"0"},
					"X-Couch-Request-ID":  {"aa1f852b27"},
					"Server":              {"CouchDB/2.1.0 (Erlang OTP/17)"},
					"Date":                {"Fri, 27 Oct 2017 19:20:04 GMT"},
					"Content-Length":      {"77"},
					"Cache-Control":       {"must-revalidate"},
				},
				ContentLength: 77,
				Body:          Body(`{"error":"bad_content_type","reason":"Content-Type must be application/json"}`),
			}, nil),
			status: http.StatusUnsupportedMediaType,
			err:    "Unsupported Media Type: Content-Type must be application/json",
		},
		{
			name: "success 2.1.0",
			query: map[string]interface{}{
				"selector": map[string]string{"_id": "foo"},
			},
			db: newTestDB(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Content-Type":        {"application/json"},
					"X-CouchDB-Body-Time": {"0"},
					"X-Couch-Request-ID":  {"a0884508d8"},
					"Server":              {"CouchDB/2.1.0 (Erlang OTP/17)"},
					"Date":                {"Fri, 27 Oct 2017 19:20:04 GMT"},
					"Transfer-Encoding":   {"chunked"},
					"Cache-Control":       {"must-revalidate"},
				},
				Body: Body(`{"docs":[
{"_id":"foo","_rev":"2-f5d2de1376388f1b54d93654df9dc9c7","_attachments":{"foo.txt":{"content_type":"text/plain","revpos":2,"digest":"md5-ENGoH7oK8V9R3BMnfDHZmw==","length":13,"stub":true}}}
]}`),
			}, nil),
		},
		{
			name: "partitioned request",
			db:   newTestDB(nil, errors.New("expected")),
			opts: map[string]interface{}{
				OptionPartition: "x2",
			},
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_partition/x2/_find"?: expected`,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Find(context.Background(), test.query, test.opts)
			testy.StatusErrorRE(t, test.err, test.status, err)
			if _, ok := result.(*rows); !ok {
				t.Errorf("Unexpected type returned: %t", result)
			}
		})
	}
}
