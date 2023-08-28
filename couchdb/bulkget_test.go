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
	"unicode"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/google/go-cmp/cmp"
)

func TestBulkGet(t *testing.T) {
	type tst struct {
		db      *db
		docs    []driver.BulkGetReference
		options map[string]interface{}
		status  int
		err     string

		rowStatus int
		rowErr    string

		expected *driver.Row
	}
	tests := testy.NewTable()
	tests.Add("network error", tst{
		db: &db{
			client: newTestClient(nil, errors.New("random network error")),
		},
		status: http.StatusBadGateway,
		err:    `^Post "?http://example.com/_bulk_get"?: random network error$`,
	})
	tests.Add("valid document", tst{
		db: &db{
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: io.NopCloser(strings.NewReader(removeSpaces(`{
	  "results": [
	    {
	      "id": "foo",
	      "docs": [
	        {
	          "ok": {
	            "_id": "foo",
	            "_rev": "4-753875d51501a6b1883a9d62b4d33f91",
	            "value": "this is foo"
	          }
	        }
	      ]
	    }
	]`))),
			}, nil),
			dbName: "xxx",
		},
		expected: &driver.Row{
			ID:  "foo",
			Doc: strings.NewReader(`{"_id":"foo","_rev":"4-753875d51501a6b1883a9d62b4d33f91","value":"thisisfoo"}`),
		},
	})
	tests.Add("invalid id", tst{
		db: &db{
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Body:       io.NopCloser(strings.NewReader(`{"results": [{"id": "", "docs": [{"error":{"id":"","rev":null,"error":"illegal_docid","reason":"Document id must not be empty"}}]}]}`)),
			}, nil),
			dbName: "xxx",
		},
		docs: []driver.BulkGetReference{{ID: ""}},
		expected: &driver.Row{
			Error: &BulkGetError{
				ID:     "",
				Rev:    "",
				Err:    "illegal_docid",
				Reason: "Document id must not be empty",
			},
		},
	})
	tests.Add("not found", tst{
		db: &db{
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				ProtoMajor: 1,
				ProtoMinor: 1,
				Body:       io.NopCloser(strings.NewReader(`{"results": [{"id": "asdf", "docs": [{"error":{"id":"asdf","rev":"1-xxx","error":"not_found","reason":"missing"}}]}]}`)),
			}, nil),
			dbName: "xxx",
		},
		docs: []driver.BulkGetReference{{ID: ""}},
		expected: &driver.Row{
			ID: "asdf",
			Error: &BulkGetError{
				ID:     "asdf",
				Rev:    "1-xxx",
				Err:    "not_found",
				Reason: "missing",
			},
		},
	})
	tests.Add("revs", tst{
		db: &db{
			client: newCustomClient(func(r *http.Request) (*http.Response, error) {
				revs := r.URL.Query().Get("revs")
				if revs != "true" {
					return nil, errors.New("Expected revs=true")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					ProtoMajor: 1,
					ProtoMinor: 1,
					Body:       io.NopCloser(strings.NewReader(`{"results": [{"id": "test1", "docs": [{"ok":{"_id":"test1","_rev":"4-8158177eb5931358b3ddaadd6377cf00","moo":123,"oink":true,"_revisions":{"start":4,"ids":["8158177eb5931358b3ddaadd6377cf00","1c08032eef899e52f35cbd1cd5f93826","e22bea278e8c9e00f3197cb2edee8bf4","7d6ff0b102072755321aa0abb630865a"]},"_attachments":{"foo.txt":{"content_type":"text/plain","revpos":2,"digest":"md5-WiGw80mG3uQuqTKfUnIZsg==","length":9,"stub":true}}}}]}]}`)),
				}, nil
			}),
			dbName: "xxx",
		},
		options: map[string]interface{}{
			"revs": true,
		},
		expected: &driver.Row{
			ID:  "test1",
			Doc: strings.NewReader(`{"_id":"test1","_rev":"4-8158177eb5931358b3ddaadd6377cf00","moo":123,"oink":true,"_revisions":{"start":4,"ids":["8158177eb5931358b3ddaadd6377cf00","1c08032eef899e52f35cbd1cd5f93826","e22bea278e8c9e00f3197cb2edee8bf4","7d6ff0b102072755321aa0abb630865a"]},"_attachments":{"foo.txt":{"content_type":"text/plain","revpos":2,"digest":"md5-WiGw80mG3uQuqTKfUnIZsg==","length":9,"stub":true}}}`),
		},
	})
	tests.Add("request", func(t *testing.T) interface{} {
		return tst{
			db: &db{
				client: newCustomClient(func(r *http.Request) (*http.Response, error) {
					defer r.Body.Close() // nolint:errcheck
					if d := testy.DiffAsJSON(testy.Snapshot(t), r.Body); d != nil {
						return nil, fmt.Errorf("Unexpected request: %s", d)
					}
					return nil, errors.New("success")
				}),
				dbName: "xxx",
			},
			docs: []driver.BulkGetReference{
				{ID: "foo"},
				{ID: "bar"},
			},
			status: 502,
			err:    "success",
		}
	})

	tests.Run(t, func(t *testing.T, test tst) {
		rows, err := test.db.BulkGet(context.Background(), test.docs, test.options)
		testy.StatusErrorRE(t, test.err, test.status, err)

		row := new(driver.Row)
		err = rows.Next(row)
		defer rows.Close() // nolint: errcheck
		testy.StatusError(t, test.rowErr, test.rowStatus, err)

		if d := rowsDiff(test.expected, row); d != "" {
			t.Error(d)
		}
	})
}

type row struct {
	ID    string
	Key   string
	Value string
	Doc   string
	Error string
}

func driverRow2row(r *driver.Row) *row {
	var value, doc []byte
	if r.Value != nil {
		value, _ = io.ReadAll(r.Value)
	}
	if r.Doc != nil {
		doc, _ = io.ReadAll(r.Doc)
	}
	var err string
	if r.Error != nil {
		err = r.Error.Error()
	}
	return &row{
		ID:    r.ID,
		Key:   string(r.Key),
		Value: string(value),
		Doc:   string(doc),
		Error: err,
	}
}

func rowsDiff(got, want *driver.Row) string {
	return cmp.Diff(driverRow2row(want), driverRow2row(got))
}

var bulkGetInput = `
{
  "results": [
    {
      "id": "foo",
      "docs": [
        {
          "ok": {
            "_id": "foo",
            "_rev": "4-753875d51501a6b1883a9d62b4d33f91",
            "value": "this is foo",
            "_revisions": {
              "start": 4,
              "ids": [
                "753875d51501a6b1883a9d62b4d33f91",
                "efc54218773c6acd910e2e97fea2a608",
                "2ee767305024673cfb3f5af037cd2729",
                "4a7e4ae49c4366eaed8edeaea8f784ad"
              ]
            }
          }
        }
      ]
    },
    {
      "id": "foo",
      "docs": [
        {
          "ok": {
            "_id": "foo",
            "_rev": "1-4a7e4ae49c4366eaed8edeaea8f784ad",
            "value": "this is the first revision of foo",
            "_revisions": {
              "start": 1,
              "ids": [
                "4a7e4ae49c4366eaed8edeaea8f784ad"
              ]
            }
          }
        }
      ]
    },
    {
      "id": "bar",
      "docs": [
        {
          "ok": {
            "_id": "bar",
            "_rev": "2-9b71d36dfdd9b4815388eb91cc8fb61d",
            "baz": true,
            "_revisions": {
              "start": 2,
              "ids": [
                "9b71d36dfdd9b4815388eb91cc8fb61d",
                "309651b95df56d52658650fb64257b97"
              ]
            }
          }
        }
      ]
    },
    {
      "id": "baz",
      "docs": [
        {
          "error": {
            "id": "baz",
            "rev": "undefined",
            "error": "not_found",
            "reason": "missing"
          }
        }
      ]
    }
  ]
}
`

func TestGetBulkRowsIterator(t *testing.T) {
	type result struct {
		ID  string
		Err string
	}
	expected := []result{
		{ID: "foo"},
		{ID: "foo"},
		{ID: "bar"},
		{ID: "baz", Err: "not_found: missing"},
	}
	results := []result{}
	rows := newBulkGetRows(context.TODO(), io.NopCloser(strings.NewReader(bulkGetInput)))
	var count int
	for {
		row := &driver.Row{}
		err := rows.Next(row)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatalf("Next() failed: %s", err)
		}
		results = append(results, result{
			ID: row.ID,
			Err: func() string {
				if row.Error == nil {
					return ""
				}
				return row.Error.Error()
			}(),
		})
		if count++; count > 10 {
			t.Fatalf("Ran too many iterations.")
		}
	}
	if d := testy.DiffInterface(expected, results); d != nil {
		t.Error(d)
	}
	if expected := 4; count != expected {
		t.Errorf("Expected %d rows, got %d", expected, count)
	}
	if err := rows.Next(&driver.Row{}); err != io.EOF {
		t.Errorf("Calling Next() after end returned unexpected error: %s", err)
	}
	if err := rows.Close(); err != nil {
		t.Errorf("Error closing rows iterator: %s", err)
	}
}

func removeSpaces(in string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return -1
		}
		return r
	}, in)
}

func TestDecodeBulkResult(t *testing.T) {
	type tst struct {
		input    string
		err      string
		expected bulkResult
	}
	tests := testy.NewTable()
	tests.Add("real example", tst{
		input: removeSpaces(`{
      "id": "test1",
      "docs": [
        {
          "ok": {
            "_id": "test1",
            "_rev": "3-1c08032eef899e52f35cbd1cd5f93826",
            "moo": 123,
            "oink": false,
            "_attachments": {
              "foo.txt": {
                "content_type": "text/plain",
                "revpos": 2,
                "digest": "md5-WiGw80mG3uQuqTKfUnIZsg==",
                "length": 9,
                "stub": true
              }
            }
          }
        }
      ]
    }`),
		expected: bulkResult{
			ID: "test1",
			Docs: []bulkResultDoc{{
				Doc: json.RawMessage(`{"_id":"test1","_rev":"3-1c08032eef899e52f35cbd1cd5f93826","moo":123,"oink":false,"_attachments":{"foo.txt":{"content_type":"text/plain","revpos":2,"digest":"md5-WiGw80mG3uQuqTKfUnIZsg==","length":9,"stub":true}}}`),
			}},
		},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		var result bulkResult
		err := json.Unmarshal([]byte(test.input), &result)
		testy.Error(t, test.err, err)
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}
