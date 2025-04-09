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
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestAllDocs(t *testing.T) {
	t.Run("standard", func(t *testing.T) {
		db := newTestDB(nil, errors.New("test error"))
		_, err := db.AllDocs(context.Background(), mock.NilOption)
		if !testy.ErrorMatchesRE(`Get "?http://example.com/testdb/_all_docs"?: test error`, err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})

	t.Run("partitioned", func(t *testing.T) {
		db := newTestDB(nil, errors.New("test error"))
		_, err := db.AllDocs(context.Background(), OptionPartition("a1"))
		if !testy.ErrorMatchesRE(`Get "?http://example.com/testdb/_partition/a1/_all_docs"?: test error`, err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}

func TestDesignDocs(t *testing.T) {
	db := newTestDB(nil, errors.New("test error"))
	_, err := db.DesignDocs(context.Background(), mock.NilOption)
	if !testy.ErrorMatchesRE(`Get "?http://example.com/testdb/_design_docs"?: test error`, err) {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestLocalDocs(t *testing.T) {
	db := newTestDB(nil, errors.New("test error"))
	_, err := db.LocalDocs(context.Background(), mock.NilOption)
	if !testy.ErrorMatchesRE(`Get "?http://example.com/testdb/_local_docs"?: test error`, err) {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestQuery(t *testing.T) {
	t.Run("standard", func(t *testing.T) {
		db := newTestDB(nil, errors.New("test error"))
		_, err := db.Query(context.Background(), "ddoc", "view", mock.NilOption)
		if !testy.ErrorMatchesRE(`Get "?http://example.com/testdb/_design/ddoc/_view/view"?: test error`, err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})
	t.Run("partitioned", func(t *testing.T) {
		db := newTestDB(nil, errors.New("test error"))
		_, err := db.Query(context.Background(), "ddoc", "view", OptionPartition("a2"))
		if !testy.ErrorMatchesRE(`Get "?http://example.com/testdb/_partition/a2/_design/ddoc/_view/view"?: test error`, err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}

type Attachment struct {
	Filename    string
	ContentType string
	Size        int64
	Content     string
}

func TestGet(t *testing.T) {
	type tt struct {
		db          *db
		id          string
		options     kivik.Option
		doc         *driver.Document
		expected    string
		attachments []*Attachment
		status      int
		err         string
	}

	tests := testy.NewTable()
	tests.Add("missing doc ID", tt{
		status: http.StatusBadRequest,
		err:    "kivik: docID required",
	})
	tests.Add("invalid options", tt{
		id:      "foo",
		options: kivik.Param("foo", make(chan int)),
		status:  http.StatusBadRequest,
		err:     "kivik: invalid type chan int for options",
	})
	tests.Add("network failure", tt{
		id:     "foo",
		db:     newTestDB(nil, errors.New("net error")),
		status: http.StatusBadGateway,
		err:    `Get "?http://example.com/testdb/foo"?: net error`,
	})
	tests.Add("error response", tt{
		id: "foo",
		db: newTestDB(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       Body(""),
		}, nil),
		status: http.StatusBadRequest,
		err:    "Bad Request",
	})
	tests.Add("status OK", tt{
		id: "foo",
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": {typeJSON},
				"ETag":         {`"12-xxx"`},
			},
			ContentLength: 13,
			Body:          Body(`{"foo":"bar"}`),
		}, nil),
		doc: &driver.Document{
			Rev: "12-xxx",
		},
		expected: `{"foo":"bar"}`,
	})
	tests.Add("If-None-Match", tt{
		db: newCustomDB(func(req *http.Request) (*http.Response, error) {
			if err := consume(req.Body); err != nil {
				return nil, err
			}
			if inm := req.Header.Get("If-None-Match"); inm != `"foo"` {
				return nil, fmt.Errorf(`If-None-Match: %s != "foo"`, inm)
			}
			return nil, errors.New("success")
		}),
		id:      "foo",
		options: OptionIfNoneMatch("foo"),
		status:  http.StatusBadGateway,
		err:     `Get "?http://example.com/testdb/foo"?: success`,
	})
	tests.Add("invalid content type in response", tt{
		id: "foo",
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": {"image/jpeg"},
				"ETag":         {`"12-xxx"`},
			},
			ContentLength: 13,
			Body:          Body("some response"),
		}, nil),
		status: http.StatusBadGateway,
		err:    "kivik: invalid content type in response: image/jpeg",
	})
	tests.Add("invalid content type header", tt{
		id: "foo",
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": {"cow; =moo"},
				"ETag":         {`"12-xxx"`},
			},
			ContentLength: 13,
			Body:          Body("some response"),
		}, nil),
		status: http.StatusBadGateway,
		err:    "mime: invalid media parameter",
	})
	tests.Add("missing multipart boundary", tt{
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Type": {typeMPRelated},
				"ETag":         {`"12-xxx"`},
			},
			ContentLength: 13,
			Body:          Body("some response"),
		}, nil),
		id:     "foo",
		status: http.StatusBadGateway,
		err:    "kivik: boundary missing for multipart/related response",
	})
	tests.Add("no multipart data", tt{
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Length": {"538"},
				"Content-Type":   {`multipart/related; boundary="e89b3e29388aef23453450d10e5aaed0"`},
				"Date":           {"Sat, 28 Sep 2013 08:08:22 GMT"},
				"ETag":           {`"2-c1c6c44c4bc3c9344b037c8690468605"`},
				"ServeR":         {"CouchDB (Erlang OTP)"},
			},
			ContentLength: 538,
			Body:          Body(`bogus data`),
		}, nil),
		id:      "foo",
		options: kivik.IncludeDocs(),
		status:  http.StatusBadGateway,
		err:     "multipart: NextPart: EOF",
	})
	tests.Add("incomplete multipart data", tt{
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Length": {"538"},
				"Content-Type":   {`multipart/related; boundary="e89b3e29388aef23453450d10e5aaed0"`},
				"Date":           {"Sat, 28 Sep 2013 08:08:22 GMT"},
				"ETag":           {`"2-c1c6c44c4bc3c9344b037c8690468605"`},
				"ServeR":         {"CouchDB (Erlang OTP)"},
			},
			ContentLength: 538,
			Body: Body(`--e89b3e29388aef23453450d10e5aaed0
				bogus data`),
		}, nil),
		id:      "foo",
		options: kivik.IncludeDocs(),
		status:  http.StatusBadGateway,
		err:     "malformed MIME header (initial )?line:.*bogus data",
	})
	tests.Add("multipart accept header", tt{
		db: newCustomDB(func(r *http.Request) (*http.Response, error) {
			expected := "multipart/mixed, multipart/related, application/json"
			if accept := r.Header.Get("Accept"); accept != expected {
				return nil, fmt.Errorf("Unexpected Accept header: %s", accept)
			}
			return nil, errors.New("not an error")
		}),
		id:     "foo",
		status: http.StatusBadGateway,
		err:    "not an error",
	})
	tests.Add("disable multipart accept header", tt{
		db: newCustomDB(func(r *http.Request) (*http.Response, error) {
			expected := "application/json"
			if accept := r.Header.Get("Accept"); accept != expected {
				return nil, fmt.Errorf("Unexpected Accept header: %s", accept)
			}
			return nil, errors.New("not an error")
		}),
		options: OptionNoMultipartGet(),
		id:      "foo",
		status:  http.StatusBadGateway,
		err:     "not an error",
	})
	tests.Add("multipart attachments", tt{
		// response borrowed from http://docs.couchdb.org/en/2.1.1/api/document/common.html#efficient-multiple-attachments-retrieving
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Length": {"538"},
				"Content-Type":   {`multipart/related; boundary="e89b3e29388aef23453450d10e5aaed0"`},
				"Date":           {"Sat, 28 Sep 2013 08:08:22 GMT"},
				"ETag":           {`"2-c1c6c44c4bc3c9344b037c8690468605"`},
				"ServeR":         {"CouchDB (Erlang OTP)"},
			},
			ContentLength: 538,
			Body: Body(`--e89b3e29388aef23453450d10e5aaed0
Content-Type: application/json

{"_id":"secret","_rev":"2-c1c6c44c4bc3c9344b037c8690468605","_attachments":{"recipe.txt":{"content_type":"text/plain","revpos":2,"digest":"md5-HV9aXJdEnu0xnMQYTKgOFA==","length":86,"follows":true}}}
--e89b3e29388aef23453450d10e5aaed0
Content-Disposition: attachment; filename="recipe.txt"
Content-Type: text/plain
Content-Length: 86

1. Take R
2. Take E
3. Mix with L
4. Add some A
5. Serve with X

--e89b3e29388aef23453450d10e5aaed0--`),
		}, nil),
		id:      "foo",
		options: kivik.IncludeDocs(),
		doc: &driver.Document{
			Rev: "2-c1c6c44c4bc3c9344b037c8690468605",
			Attachments: &multipartAttachments{
				meta: map[string]attMeta{
					"recipe.txt": {
						Follows:     true,
						ContentType: "text/plain",
						Size:        func() *int64 { x := int64(86); return &x }(),
					},
				},
			},
		},
		expected: `{"_id":"secret","_rev":"2-c1c6c44c4bc3c9344b037c8690468605","_attachments":{"recipe.txt":{"content_type":"text/plain","revpos":2,"digest":"md5-HV9aXJdEnu0xnMQYTKgOFA==","length":86,"follows":true}}}`,
		attachments: []*Attachment{
			{
				Filename:    "recipe.txt",
				Size:        86,
				ContentType: "text/plain",
				Content:     "1. Take R\n2. Take E\n3. Mix with L\n4. Add some A\n5. Serve with X\n",
			},
		},
	})
	tests.Add("multipart attachments, doc content length", tt{
		// response borrowed from http://docs.couchdb.org/en/2.1.1/api/document/common.html#efficient-multiple-attachments-retrieving
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Header: http.Header{
				"Content-Length": {"558"},
				"Content-Type":   {`multipart/related; boundary="e89b3e29388aef23453450d10e5aaed0"`},
				"Date":           {"Sat, 28 Sep 2013 08:08:22 GMT"},
				"ETag":           {`"2-c1c6c44c4bc3c9344b037c8690468605"`},
				"ServeR":         {"CouchDB (Erlang OTP)"},
			},
			ContentLength: 558,
			Body: Body(`--e89b3e29388aef23453450d10e5aaed0
Content-Type: application/json
Content-Length: 199

{"_id":"secret","_rev":"2-c1c6c44c4bc3c9344b037c8690468605","_attachments":{"recipe.txt":{"content_type":"text/plain","revpos":2,"digest":"md5-HV9aXJdEnu0xnMQYTKgOFA==","length":86,"follows":true}}}
--e89b3e29388aef23453450d10e5aaed0
Content-Disposition: attachment; filename="recipe.txt"
Content-Type: text/plain
Content-Length: 86

1. Take R
2. Take E
3. Mix with L
4. Add some A
5. Serve with X

--e89b3e29388aef23453450d10e5aaed0--`),
		}, nil),
		id:      "foo",
		options: kivik.IncludeDocs(),
		doc: &driver.Document{
			Rev: "2-c1c6c44c4bc3c9344b037c8690468605",
			Attachments: &multipartAttachments{
				meta: map[string]attMeta{
					"recipe.txt": {
						Follows:     true,
						ContentType: "text/plain",
						Size:        func() *int64 { x := int64(86); return &x }(),
					},
				},
			},
		},
		expected: `{"_id":"secret","_rev":"2-c1c6c44c4bc3c9344b037c8690468605","_attachments":{"recipe.txt":{"content_type":"text/plain","revpos":2,"digest":"md5-HV9aXJdEnu0xnMQYTKgOFA==","length":86,"follows":true}}}`,
		attachments: []*Attachment{
			{
				Filename:    "recipe.txt",
				Size:        86,
				ContentType: "text/plain",
				Content:     "1. Take R\n2. Take E\n3. Mix with L\n4. Add some A\n5. Serve with X\n",
			},
		},
	})
	tests.Add("bug268 - complex id", func(*testing.T) interface{} {
		return tt{
			db: newCustomDB(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("success")
			}),
			id:     "http://example.com/",
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/testdb/http%3A%2F%2Fexample\.com%2F"?: success`,
		}
	})
	tests.Add("plus sign", func(*testing.T) interface{} {
		return tt{
			db: newCustomDB(func(*http.Request) (*http.Response, error) {
				return nil, errors.New("success")
			}),
			id:     "2020-01-30T13:33:00.00+05:30|kl",
			status: http.StatusBadGateway,
			err:    `^Get "?http://example.com/testdb/2020-01-30T13%3A33%3A00\.00%2B05%3A30%7Ckl"?: success$`,
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		result, err := tt.db.Get(context.Background(), tt.id, opts)
		if !testy.ErrorMatchesRE(tt.err, err) {
			t.Errorf("Unexpected error: \n Got: %s\nWant: /%s/", err, tt.err)
		}
		if err != nil {
			return
		}

		if d := testy.DiffAsJSON([]byte(tt.expected), result.Body); d != nil {
			t.Errorf("Unexpected result: %s", d)
		}
		attachments := rowAttachments(t, result.Attachments)

		_ = result.Body.Close()
		result.Body = nil // Determinism
		if d := testy.DiffInterface(tt.doc, result); d != nil {
			t.Errorf("Unexpected doc:\n%s", d)
		}
		if d := testy.DiffInterface(tt.attachments, attachments); d != nil {
			t.Errorf("Unexpected attachments:\n%s", d)
		}
	})
}

func rowAttachments(t *testing.T, atts driver.Attachments) []*Attachment {
	t.Helper()
	var attachments []*Attachment
	if atts != nil {
		att := new(driver.Attachment)
		for {
			if err := atts.Next(att); err != nil {
				if err != io.EOF {
					t.Fatal(err)
				}
				break
			}
			content, e := io.ReadAll(att.Content)
			if e != nil {
				t.Fatal(e)
			}
			attachments = append(attachments, &Attachment{
				Filename:    att.Filename,
				ContentType: att.ContentType,
				Size:        att.Size,
				Content:     string(content),
			})
		}
		atts.(*multipartAttachments).content = nil // Determinism
		atts.(*multipartAttachments).mpReader = nil
	}
	return attachments
}

func TestOpenRevs(t *testing.T) {
	type rowResult struct {
		ID    string
		Rev   string
		Error string
	}
	type tt struct {
		db      *db
		id      string
		revs    []string
		options kivik.Option
		want    []rowResult
		err     string
	}

	tests := testy.NewTable()
	tests.Add("open_revs", func(*testing.T) interface{} {
		return tt{
			db: newCustomDB(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{`multipart/mixed; boundary="ea68bec945fd9dece3e826462c5604e8"`},
					},
					Body: Body(`--ea68bec945fd9dece3e826462c5604e8
Content-Type: application/json

{"_id":"bar","_rev":"2-e2a6df12e36615e8def0bb38bb17b48d","foo":123}
--ea68bec945fd9dece3e826462c5604e8--
`),
				}, nil
			}),
			id: "bar",
			want: []rowResult{
				{
					ID:  "bar",
					Rev: "2-e2a6df12e36615e8def0bb38bb17b48d",
				},
			},
		}
	})
	tests.Add("open_revs with multiple revs", func(*testing.T) interface{} {
		return tt{
			db: newCustomDB(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Content-Type": []string{`multipart/mixed; boundary="7b1596fc4940bc1be725ad67f11ec1c4"`},
					},
					Body: Body(`--7b1596fc4940bc1be725ad67f11ec1c4
Content-Type: application/json

{
	"_id": "SpaghettiWithMeatballs",
	"_rev": "1-917fa23",
	"_revisions": {
		"ids": [
			"917fa23"
		],
		"start": 1
	},
	"description": "An Italian-American delicious dish",
	"ingredients": [
		"spaghetti",
		"tomato sauce",
		"meatballs"
	],
	"name": "Spaghetti with meatballs"
}
--7b1596fc4940bc1be725ad67f11ec1c4
Content-Type: multipart/related; boundary="a81a77b0ca68389dda3243a43ca946f2"

--a81a77b0ca68389dda3243a43ca946f2
Content-Type: application/json

{
	"_attachments": {
		"recipe.txt": {
			"content_type": "text/plain",
			"digest": "md5-R5CrCb6fX10Y46AqtNn0oQ==",
			"follows": true,
			"length": 87,
			"revpos": 7
		}
	},
	"_id": "SpaghettiWithMeatballs",
	"_rev": "7-474f12e",
	"_revisions": {
		"ids": [
			"474f12e",
			"5949cfc",
			"00ecbbc",
			"fc997b6",
			"3552c87",
			"404838b",
			"5defd9d",
			"dc1e4be"
		],
		"start": 7
	},
	"description": "An Italian-American delicious dish",
	"ingredients": [
		"spaghetti",
		"tomato sauce",
		"meatballs",
		"love"
	],
	"name": "Spaghetti with meatballs"
}
--a81a77b0ca68389dda3243a43ca946f2
Content-Disposition: attachment; filename="recipe.txt"
Content-Type: text/plain
Content-Length: 87

1. Cook spaghetti
2. Cook meetballs
3. Mix them
4. Add tomato sauce
5. ...
6. PROFIT!

--a81a77b0ca68389dda3243a43ca946f2--
--7b1596fc4940bc1be725ad67f11ec1c4
Content-Type: application/json; error="true"

{"missing":"3-6bcedf1"}
--7b1596fc4940bc1be725ad67f11ec1c4--`),
				}, nil
			}),
			id: "bar",
			want: []rowResult{
				{
					ID:  "bar",
					Rev: "1-917fa23",
				},
				{
					ID:  "bar",
					Rev: "7-474f12e",
				},
				{
					ID:    "bar",
					Rev:   "3-6bcedf1",
					Error: "missing",
				},
			},
		}
	})
	tests.Add("not found", func(*testing.T) interface{} {
		return tt{
			db: newCustomDB(func(*http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusNotFound,
					Header: http.Header{
						"Content-Type": []string{`application/json`},
					},
					Body: Body(`{"error":"not_found","reason":"missing"}`),
				}, nil
			}),
			id:  "bar",
			err: "Not Found",
		}
	})
	tests.Run(t, func(t *testing.T, tt tt) {
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		rows, err := tt.db.OpenRevs(context.Background(), tt.id, tt.revs, opts)
		var errMsg string
		if err != nil {
			errMsg = err.Error()
		}
		if errMsg != tt.err {
			t.Errorf("Unexpected error: %s", err)
		}
		if errMsg != "" {
			return
		}

		got := []rowResult{}
		for i := 0; ; i++ {
			row := new(driver.Row)
			if err := rows.Next(row); err != nil {
				if err == io.EOF {
					break
				}
				t.Fatal(err)
			}
			row.Doc = nil // Determinism
			row.Attachments = nil
			got = append(got, rowResult{
				ID:  row.ID,
				Rev: row.Rev,
				Error: func() string {
					if row.Error != nil {
						return row.Error.Error()
					}
					return ""
				}(),
			})
		}
		if d := testy.DiffInterface(tt.want, got); d != nil {
			t.Errorf("Unexpected result: %s", d)
		}
	})
}

func TestCreateDoc(t *testing.T) {
	tests := []struct {
		name    string
		db      *db
		doc     interface{}
		options kivik.Option
		id, rev string
		status  int
		err     string
	}{
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("foo error")),
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb"?: foo error`,
		},
		{
			name:   "invalid doc",
			doc:    make(chan int),
			db:     newTestDB(nil, errors.New("")),
			status: http.StatusBadRequest,
			err:    `Post "?http://example.com/testdb"?: json: unsupported type: chan int`,
		},
		{
			name: "error response",
			doc:  map[string]interface{}{"foo": "bar"},
			db: newTestDB(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil),
			status: http.StatusBadRequest,
			err:    "Bad Request",
		},
		{
			name: "invalid JSON response",
			doc:  map[string]interface{}{"foo": "bar"},
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("invalid json")),
			}, nil),
			status: http.StatusBadGateway,
			err:    "invalid character 'i' looking for beginning of value",
		},
		{
			name: "success, 1.6.1",
			doc:  map[string]interface{}{"foo": "bar"},
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Location":       {"http://localhost:5984/foo/43734cf3ce6d5a37050c050bb600006b"},
					"ETag":           {`"1-4c6114c65e295552ab1019e2b046b10e"`},
					"Date":           {"Wed, 25 Oct 2017 10:38:38 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"95"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: io.NopCloser(strings.NewReader(`{"ok":true,"id":"43734cf3ce6d5a37050c050bb600006b","rev":"1-4c6114c65e295552ab1019e2b046b10e"}
`)),
			}, nil),
			id:  "43734cf3ce6d5a37050c050bb600006b",
			rev: "1-4c6114c65e295552ab1019e2b046b10e",
		},
		{
			name:    "batch mode",
			db:      newTestDB(nil, errors.New("success")),
			doc:     map[string]string{"foo": "bar"},
			options: kivik.Param("batch", "ok"),
			status:  http.StatusBadGateway,
			err:     `^Post "?http://example.com/testdb\?batch=ok"?: success$`,
		},
		{
			name: "full commit",
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				if err := consume(req.Body); err != nil {
					return nil, err
				}
				if fullCommit := req.Header.Get("X-Couch-Full-Commit"); fullCommit != "true" {
					return nil, errors.New("X-Couch-Full-Commit not true")
				}
				return nil, errors.New("success")
			}),
			options: OptionFullCommit(),
			status:  http.StatusBadGateway,
			err:     `Post "?http://example.com/testdb"?: success`,
		},
		{
			name:    "invalid options",
			db:      &db{},
			options: kivik.Param("foo", make(chan int)),
			status:  http.StatusBadRequest,
			err:     "kivik: invalid type chan int for options",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := test.options
			if opts == nil {
				opts = mock.NilOption
			}
			id, rev, err := test.db.CreateDoc(context.Background(), test.doc, opts)
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if test.id != id || test.rev != rev {
				t.Errorf("Unexpected results: ID=%s rev=%s", id, rev)
			}
		})
	}
}

func TestOptionsToParams(t *testing.T) {
	type otpTest struct {
		Name     string
		Input    map[string]interface{}
		Expected url.Values
		Error    string
	}
	tests := []otpTest{
		{
			Name:  "Unmarshalable key",
			Input: map[string]interface{}{"key": make(chan int)},
			Error: "json: unsupported type: chan int",
		},
		{
			Name:     "String",
			Input:    map[string]interface{}{"foo": "bar"},
			Expected: map[string][]string{"foo": {"bar"}},
		},
		{
			Name:     "StringSlice",
			Input:    map[string]interface{}{"foo": []string{"bar", "baz"}},
			Expected: map[string][]string{"foo": {"bar", "baz"}},
		},
		{
			Name:     "Bool",
			Input:    map[string]interface{}{"foo": true},
			Expected: map[string][]string{"foo": {"true"}},
		},
		{
			Name:     "Int",
			Input:    map[string]interface{}{"foo": 123},
			Expected: map[string][]string{"foo": {"123"}},
		},
		{
			Name:  "Error",
			Input: map[string]interface{}{"foo": []byte("foo")},
			Error: "kivik: invalid type []uint8 for options",
		},
	}
	for _, test := range tests {
		func(test otpTest) {
			t.Run(test.Name, func(t *testing.T) {
				params, err := optionsToParams(test.Input)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.Error {
					t.Errorf("Error\n\tExpected: %s\n\t  Actual: %s\n", test.Error, msg)
				}
				if d := testy.DiffInterface(test.Expected, params); d != nil {
					t.Errorf("Params not as expected:\n%s\n", d)
				}
			})
		}(test)
	}
}

func TestCompact(t *testing.T) {
	tests := []struct {
		name   string
		db     *db
		status int
		err    string
	}{
		{
			name:   "net error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_compact"?: net error`,
		},
		{
			name: "1.6.1",
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				if ct, _, _ := mime.ParseMediaType(req.Header.Get("Content-Type")); ct != typeJSON {
					return nil, fmt.Errorf("Expected Content-Type: application/json, got %s", ct)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
						"Date":           {"Thu, 26 Oct 2017 13:07:52 GMT"},
						"Content-Type":   {"text/plain; charset=utf-8"},
						"Content-Length": {"12"},
						"Cache-Control":  {"must-revalidate"},
					},
					Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
				}, nil
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.Compact(context.Background())
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestCompactView(t *testing.T) {
	tests := []struct {
		name   string
		db     *db
		id     string
		status int
		err    string
	}{
		{
			name:   "no ddoc",
			status: http.StatusBadRequest,
			err:    "kivik: ddocID required",
		},
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			id:     "foo",
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_compact/foo"?: net error`,
		},
		{
			name: "1.6.1",
			id:   "foo",
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				if ct, _, _ := mime.ParseMediaType(req.Header.Get("Content-Type")); ct != typeJSON {
					return nil, fmt.Errorf("Expected Content-Type: application/json, got %s", ct)
				}
				return &http.Response{
					StatusCode: http.StatusAccepted,
					Header: http.Header{
						"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
						"Date":           {"Thu, 26 Oct 2017 13:07:52 GMT"},
						"Content-Type":   {"text/plain; charset=utf-8"},
						"Content-Length": {"12"},
						"Cache-Control":  {"must-revalidate"},
					},
					Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
				}, nil
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.CompactView(context.Background(), test.id)
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestViewCleanup(t *testing.T) {
	tests := []struct {
		name   string
		db     *db
		status int
		err    string
	}{
		{
			name:   "net error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_view_cleanup"?: net error`,
		},
		{
			name: "1.6.1",
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				if ct, _, _ := mime.ParseMediaType(req.Header.Get("Content-Type")); ct != typeJSON {
					return nil, fmt.Errorf("Expected Content-Type: application/json, got %s", ct)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
						"Date":           {"Thu, 26 Oct 2017 13:07:52 GMT"},
						"Content-Type":   {"text/plain; charset=utf-8"},
						"Content-Length": {"12"},
						"Cache-Control":  {"must-revalidate"},
					},
					Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
				}, nil
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.ViewCleanup(context.Background())
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestPut(t *testing.T) {
	type test struct {
		db      *db
		id      string
		doc     interface{}
		options kivik.Option
		rev     string
		status  int
		err     string
	}

	tests := testy.NewTable()
	tests.Add("missing docID", test{
		status: http.StatusBadRequest,
		err:    "kivik: docID required",
	})
	tests.Add("network error", test{
		id:     "foo",
		db:     newTestDB(nil, errors.New("net error")),
		status: http.StatusBadGateway,
		err:    `Put "?http://example.com/testdb/foo"?: net error`,
	})
	tests.Add("error response", test{
		id: "foo",
		db: newTestDB(&http.Response{
			StatusCode: http.StatusBadRequest,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil),
		status: http.StatusBadRequest,
		err:    "Bad Request",
	})
	tests.Add("invalid JSON response", test{
		id: "foo",
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("invalid json")),
		}, nil),
		status: http.StatusBadGateway,
		err:    "invalid character 'i' looking for beginning of value",
	})
	tests.Add("invalid document", test{
		id:  "foo",
		doc: make(chan int),
		db: newTestDB(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader("")),
		}, nil),
		status: http.StatusBadRequest,
		err:    `Put "?http://example.com/testdb/foo"?: json: unsupported type: chan int`,
	})
	tests.Add("doc created, 1.6.1", test{
		id:  "foo",
		doc: map[string]string{"foo": "bar"},
		db: newTestDB(&http.Response{
			StatusCode: http.StatusCreated,
			Header: http.Header{
				"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
				"Location":       {"http://localhost:5984/foo/foo"},
				"ETag":           {`"1-4c6114c65e295552ab1019e2b046b10e"`},
				"Date":           {"Wed, 25 Oct 2017 12:33:09 GMT"},
				"Content-Type":   {"text/plain; charset=utf-8"},
				"Content-Length": {"66"},
				"Cache-Control":  {"must-revalidate"},
			},
			Body: io.NopCloser(strings.NewReader(`{"ok":true,"id":"foo","rev":"1-4c6114c65e295552ab1019e2b046b10e"}`)),
		}, nil),
		rev: "1-4c6114c65e295552ab1019e2b046b10e",
	})
	tests.Add("full commit", test{
		db: newCustomDB(func(req *http.Request) (*http.Response, error) {
			if err := consume(req.Body); err != nil {
				return nil, err
			}
			if fullCommit := req.Header.Get("X-Couch-Full-Commit"); fullCommit != "true" {
				return nil, errors.New("X-Couch-Full-Commit not true")
			}
			return nil, errors.New("success")
		}),
		id:      "foo",
		doc:     map[string]string{"foo": "bar"},
		options: OptionFullCommit(),
		status:  http.StatusBadGateway,
		err:     `Put "?http://example.com/testdb/foo"?: success`,
	})
	tests.Add("connection refused", test{
		db: func() *db {
			c, err := chttp.New(&http.Client{}, "http://127.0.0.1:1/", mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
			return &db{
				client: &client{Client: c},
				dbName: "animals",
			}
		}(),
		id:     "cow",
		doc:    map[string]interface{}{"feet": 4},
		status: http.StatusBadGateway,
		err:    `Put "?http://127\.0\.0\.1:1/animals/cow"?: dial tcp (\[::1\]|127.0.0.1):1: (getsockopt|connect): connection refused`,
	})
	tests.Add("real database, multipart attachments", func(t *testing.T) interface{} {
		db := realDB(t)
		t.Cleanup(func() {
			if err := db.client.DestroyDB(context.Background(), db.dbName, nil); err != nil {
				t.Fatal(err)
			}
		})

		return test{
			db: db,
			id: "foo",
			doc: map[string]interface{}{
				"feet": 4,
				"_attachments": &kivik.Attachments{
					"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
				},
			},
			rev: "1-1e527110339245a3191b3f6cbea27ab1",
		}
	})

	tests.Run(t, func(t *testing.T, tt test) {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		rev, err := tt.db.Put(ctx, tt.id, tt.doc, opts)
		if d := internal.StatusErrorDiffRE(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if rev != tt.rev {
			t.Errorf("Unexpected rev: %s", rev)
		}
	})
}

func TestDelete(t *testing.T) {
	type tt struct {
		db      *db
		id      string
		options kivik.Option
		newrev  string
		status  int
		err     string
	}

	tests := testy.NewTable()
	tests.Add("no doc id", tt{
		status: http.StatusBadRequest,
		err:    "kivik: docID required",
	})
	tests.Add("no rev", tt{
		id:     "foo",
		status: http.StatusBadRequest,
		err:    "kivik: rev required",
	})
	tests.Add("network error", tt{
		id:      "foo",
		options: kivik.Rev("1-xxx"),
		db:      newTestDB(nil, errors.New("net error")),
		status:  http.StatusBadGateway,
		err:     `(Delete "?http://example.com/testdb/foo\?rev="?: )?net error`,
	})
	tests.Add("1.6.1 conflict", tt{
		id:      "43734cf3ce6d5a37050c050bb600006b",
		options: kivik.Rev("1-xxx"),
		db: newTestDB(&http.Response{
			StatusCode: 409,
			Header: http.Header{
				"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
				"Date":           {"Thu, 26 Oct 2017 13:29:06 GMT"},
				"Content-Type":   {"text/plain; charset=utf-8"},
				"Content-Length": {"58"},
				"Cache-Control":  {"must-revalidate"},
			},
			Body: io.NopCloser(strings.NewReader(`{"error":"conflict","reason":"Document update conflict."}`)),
		}, nil),
		status: http.StatusConflict,
		err:    "Conflict",
	})
	tests.Add("1.6.1 success", tt{
		id:      "43734cf3ce6d5a37050c050bb600006b",
		options: kivik.Rev("1-4c6114c65e295552ab1019e2b046b10e"),
		db: newTestDB(&http.Response{
			StatusCode: 200,
			Header: http.Header{
				"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
				"Date":           {"Thu, 26 Oct 2017 13:29:06 GMT"},
				"Content-Type":   {"text/plain; charset=utf-8"},
				"ETag":           {`"2-185ccf92154a9f24a4f4fd12233bf463"`},
				"Content-Length": {"95"},
				"Cache-Control":  {"must-revalidate"},
			},
			Body: io.NopCloser(strings.NewReader(`{"ok":true,"id":"43734cf3ce6d5a37050c050bb600006b","rev":"2-185ccf92154a9f24a4f4fd12233bf463"}`)),
		}, nil),
		newrev: "2-185ccf92154a9f24a4f4fd12233bf463",
	})
	tests.Add("batch mode", tt{
		db: newCustomDB(func(req *http.Request) (*http.Response, error) {
			if err := consume(req.Body); err != nil {
				return nil, err
			}
			if batch := req.URL.Query().Get("batch"); batch != "ok" {
				return nil, fmt.Errorf("Unexpected query batch=%s", batch)
			}
			return nil, errors.New("success")
		}),
		id: "foo",
		options: kivik.Params(map[string]interface{}{
			"batch": "ok",
			"rev":   "1-xxx",
		}),
		status: http.StatusBadGateway,
		err:    "success",
	})
	tests.Add("invalid options", tt{
		db: &db{},
		id: "foo",
		options: kivik.Params(map[string]interface{}{
			"foo": make(chan int),
			"rev": "1-xxx",
		}),
		status: http.StatusBadRequest,
		err:    "kivik: invalid type chan int for options",
	})
	tests.Add("full commit", tt{
		db: newCustomDB(func(req *http.Request) (*http.Response, error) {
			if err := consume(req.Body); err != nil {
				return nil, err
			}
			if fullCommit := req.Header.Get("X-Couch-Full-Commit"); fullCommit != "true" {
				return nil, errors.New("X-Couch-Full-Commit not true")
			}
			return nil, errors.New("success")
		}),
		id: "foo",
		options: multiOptions{
			OptionFullCommit(),
			kivik.Rev("1-xxx"),
		},
		status: http.StatusBadGateway,
		err:    "success",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		newrev, err := tt.db.Delete(context.Background(), tt.id, opts)
		if d := internal.StatusErrorDiffRE(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if newrev != tt.newrev {
			t.Errorf("Unexpected new rev: %s", newrev)
		}
	})
}

func TestFlush(t *testing.T) {
	tests := []struct {
		name   string
		db     *db
		status int
		err    string
	}{
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Post "?http://example.com/testdb/_ensure_full_commit"?: net error`,
		},
		{
			name: "1.6.1",
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				if ct, _, _ := mime.ParseMediaType(req.Header.Get("Content-Type")); ct != typeJSON {
					return nil, fmt.Errorf("Expected Content-Type: application/json, got %s", ct)
				}
				return &http.Response{
					StatusCode: 201,
					Header: http.Header{
						"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
						"Date":           {"Thu, 26 Oct 2017 13:07:52 GMT"},
						"Content-Type":   {"text/plain; charset=utf-8"},
						"Content-Length": {"53"},
						"Cache-Control":  {"must-revalidate"},
					},
					Body: io.NopCloser(strings.NewReader(`{"ok":true,"instance_start_time":"1509022681259533"}`)),
				}, nil
			}),
		},
		{
			name: "2.0.0",
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				if ct, _, _ := mime.ParseMediaType(req.Header.Get("Content-Type")); ct != typeJSON {
					return nil, fmt.Errorf("Expected Content-Type: application/json, got %s", ct)
				}
				return &http.Response{
					StatusCode: 201,
					Header: http.Header{
						"Server":              {"CouchDB/2.0.0 (Erlang OTP/17)"},
						"Date":                {"Thu, 26 Oct 2017 13:07:52 GMT"},
						"Content-Type":        {typeJSON},
						"Content-Length":      {"38"},
						"Cache-Control":       {"must-revalidate"},
						"X-Couch-Request-ID":  {"e454023cb8"},
						"X-CouchDB-Body-Time": {"0"},
					},
					Body: io.NopCloser(strings.NewReader(`{"ok":true,"instance_start_time":"0"}`)),
				}, nil
			}),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.db.Flush(context.Background())
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
		})
	}
}

type queryResult struct {
	Offset    int64
	TotalRows int64
	Warning   string
	UpdateSeq string
	Err       string
	Rows      []*driver.Row
}

func queryResultDiff(got, want queryResult) string {
	type qr struct {
		Offset    int64
		TotalRows int64
		Warning   string
		UpdateSeq string
		Err       string
		Rows      []*row
	}
	g := qr{
		Offset:    got.Offset,
		TotalRows: got.TotalRows,
		Warning:   got.Warning,
		UpdateSeq: got.UpdateSeq,
		Err:       got.Err,
		Rows:      make([]*row, len(got.Rows)),
	}
	for i, row := range got.Rows {
		g.Rows[i] = driverRow2row(row)
	}

	w := qr{
		Offset:    want.Offset,
		TotalRows: want.TotalRows,
		Warning:   want.Warning,
		UpdateSeq: want.UpdateSeq,
		Err:       want.Err,
		Rows:      make([]*row, len(want.Rows)),
	}
	for i, row := range want.Rows {
		w.Rows[i] = driverRow2row(row)
	}
	return cmp.Diff(g, w)
}

func TestRowsQuery(t *testing.T) {
	tests := []struct {
		name     string
		db       *db
		path     string
		options  kivik.Option
		expected queryResult
		status   int
		err      string
	}{
		{
			name:    "invalid options",
			path:    "_all_docs",
			options: kivik.Param("foo", make(chan int)),
			status:  http.StatusBadRequest,
			err:     "kivik: invalid type chan int for options",
		},
		{
			name:   "network error",
			path:   "_all_docs",
			db:     newTestDB(nil, errors.New("go away")),
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/testdb/_all_docs"?: go away`,
		},
		{
			name: "error response",
			path: "_all_docs",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil),
			status: http.StatusBadRequest,
			err:    "Bad Request",
		},
		{
			name: "all docs default 1.6.1",
			path: "_all_docs",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"Transfer-Encoding": {"chunked"},
					"Date":              {"Tue, 24 Oct 2017 21:17:12 GMT"},
					"Server":            {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"ETag":              {`"2MVNDK3T2PN4JUK89TKD10QDA"`},
					"Content-Type":      {"text/plain; charset=utf-8"},
					"Cache-Control":     {"must-revalidate"},
				},
				Body: io.NopCloser(strings.NewReader(`{"total_rows":3,"offset":0,"rows":[
{"id":"_design/_auth","key":"_design/_auth","value":{"rev":"1-75efcce1f083316d622d389f3f9813f7"}},
{"id":"org.couchdb.user:5wmxzru3b4i6pdmvhslq5egiye","key":"org.couchdb.user:5wmxzru3b4i6pdmvhslq5egiye","value":{"rev":"1-747e6766038164010fd0efcabd1a31dd"}},
{"id":"org.couchdb.user:zqfdn6u3cqi6pol3hslq5egiye","key":"org.couchdb.user:zqfdn6u3cqi6pol3hslq5egiye","value":{"rev":"1-4645438e6e1aa2230a1b06b5c1f5c63f"}}
]}
`)),
			}, nil),
			expected: queryResult{
				TotalRows: 3,
				Rows: []*driver.Row{
					{
						ID:    "_design/_auth",
						Key:   []byte(`"_design/_auth"`),
						Value: strings.NewReader(`{"rev":"1-75efcce1f083316d622d389f3f9813f7"}`),
					},
					{
						ID:    "org.couchdb.user:5wmxzru3b4i6pdmvhslq5egiye",
						Key:   []byte(`"org.couchdb.user:5wmxzru3b4i6pdmvhslq5egiye"`),
						Value: strings.NewReader(`{"rev":"1-747e6766038164010fd0efcabd1a31dd"}`),
					},
					{
						ID:    "org.couchdb.user:zqfdn6u3cqi6pol3hslq5egiye",
						Key:   []byte(`"org.couchdb.user:zqfdn6u3cqi6pol3hslq5egiye"`),
						Value: strings.NewReader(`{"rev":"1-4645438e6e1aa2230a1b06b5c1f5c63f"}`),
					},
				},
			},
		},
		{
			name: "all docs options 1.6.1",
			path: "/_all_docs?update_seq=true&limit=1&skip=1",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"Transfer-Encoding": {"chunked"},
					"Date":              {"Tue, 24 Oct 2017 21:17:12 GMT"},
					"Server":            {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"ETag":              {`"2MVNDK3T2PN4JUK89TKD10QDA"`},
					"Content-Type":      {"text/plain; charset=utf-8"},
					"Cache-Control":     {"must-revalidate"},
				},
				Body: io.NopCloser(strings.NewReader(`{"total_rows":3,"offset":1,"update_seq":31,"rows":[
{"id":"org.couchdb.user:5wmxzru3b4i6pdmvhslq5egiye","key":"org.couchdb.user:5wmxzru3b4i6pdmvhslq5egiye","value":{"rev":"1-747e6766038164010fd0efcabd1a31dd"}}
]}
`)),
			}, nil),
			expected: queryResult{
				TotalRows: 3,
				Offset:    1,
				UpdateSeq: "31",
				Rows: []*driver.Row{
					{
						ID:    "org.couchdb.user:5wmxzru3b4i6pdmvhslq5egiye",
						Key:   []byte(`"org.couchdb.user:5wmxzru3b4i6pdmvhslq5egiye"`),
						Value: strings.NewReader(`{"rev":"1-747e6766038164010fd0efcabd1a31dd"}`),
					},
				},
			},
		},
		{
			name: "all docs options 2.0.0, no results",
			path: "/_all_docs?update_seq=true&limit=1",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Header: map[string][]string{
					"Transfer-Encoding":  {"chunked"},
					"Date":               {"Tue, 24 Oct 2017 21:21:30 GMT"},
					"Server":             {"CouchDB/2.0.0 (Erlang OTP/17)"},
					"Content-Type":       {typeJSON},
					"Cache-Control":      {"must-revalidate"},
					"X-Couch-Request-ID": {"a9688d9335"},
					"X-Couch-Body-Time":  {"0"},
				},
				Body: io.NopCloser(strings.NewReader(`{"total_rows":1,"offset":0,"update_seq":"13-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjPlsQBJhgdA6j8QZCUy4Fv4AKLuflYiE151DRB18wmZtwCibj9u85ISgGRSPV63JSmA1NiD1bDgUJPIkCSP3xAHkCHxYDWsWQDg12MD","rows":[
{"id":"_design/_auth","key":"_design/_auth","value":{"rev":"1-75efcce1f083316d622d389f3f9813f7"}}
]}
`)),
			}, nil),
			expected: queryResult{
				TotalRows: 1,
				UpdateSeq: "13-g1AAAAEzeJzLYWBg4MhgTmHgzcvPy09JdcjLz8gvLskBCjPlsQBJhgdA6j8QZCUy4Fv4AKLuflYiE151DRB18wmZtwCibj9u85ISgGRSPV63JSmA1NiD1bDgUJPIkCSP3xAHkCHxYDWsWQDg12MD",
				Rows: []*driver.Row{
					{
						ID:    "_design/_auth",
						Key:   []byte(`"_design/_auth"`),
						Value: strings.NewReader(`{"rev":"1-75efcce1f083316d622d389f3f9813f7"}`),
					},
				},
			},
		},
		{
			name:    "all docs with keys",
			path:    "/_all_docs",
			options: kivik.Param("keys", []string{"_design/_auth", "foo"}),
			db: newCustomDB(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPost {
					t.Errorf("Unexpected method: %s", r.Method)
				}
				defer r.Body.Close() // nolint: errcheck
				if d := testy.DiffAsJSON(map[string][]string{"keys": {"_design/_auth", "foo"}}, r.Body); d != nil {
					t.Error(d)
				}
				if keys := r.URL.Query().Get("keys"); keys != "" {
					t.Error("query key 'keys' should be absent")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Transfer-Encoding":  {"chunked"},
						"Date":               {"Sat, 01 Sep 2018 19:01:30 GMT"},
						"Server":             {"CouchDB/2.2.0 (Erlang OTP/19)"},
						"Content-Type":       {typeJSON},
						"Cache-Control":      {"must-revalidate"},
						"X-Couch-Request-ID": {"24fdb3fd86"},
						"X-Couch-Body-Time":  {"0"},
					},
					Body: io.NopCloser(strings.NewReader(`{"total_rows":1,"offset":null,"rows":[
{"id":"_design/_auth","key":"_design/_auth","value":{"rev":"1-6e609020e0371257432797b4319c5829"}}
]}`)),
				}, nil
			}),
			expected: queryResult{
				TotalRows: 1,
				UpdateSeq: "",
				Rows: []*driver.Row{
					{
						ID:    "_design/_auth",
						Key:   []byte(`"_design/_auth"`),
						Value: strings.NewReader(`{"rev":"1-6e609020e0371257432797b4319c5829"}`),
					},
				},
			},
		},
		{
			name:    "all docs with endkey",
			path:    "/_all_docs",
			options: kivik.Param("endkey", []string{"foo", "bar"}),
			db: newCustomDB(func(r *http.Request) (*http.Response, error) {
				if d := testy.DiffAsJSON([]byte(`["foo","bar"]`), []byte(r.URL.Query().Get("endkey"))); d != nil {
					t.Error(d)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Transfer-Encoding":  {"chunked"},
						"Date":               {"Sat, 01 Sep 2018 19:01:30 GMT"},
						"Server":             {"CouchDB/2.2.0 (Erlang OTP/19)"},
						"Content-Type":       {typeJSON},
						"Cache-Control":      {"must-revalidate"},
						"X-Couch-Request-ID": {"24fdb3fd86"},
						"X-Couch-Body-Time":  {"0"},
					},
					Body: io.NopCloser(strings.NewReader(`{"total_rows":1,"offset":null,"rows":[
{"id":"_design/_auth","key":"_design/_auth","value":{"rev":"1-6e609020e0371257432797b4319c5829"}}
]}`)),
				}, nil
			}),
			expected: queryResult{
				TotalRows: 1,
				UpdateSeq: "",
				Rows: []*driver.Row{
					{
						ID:    "_design/_auth",
						Key:   []byte(`"_design/_auth"`),
						Value: strings.NewReader(`{"rev":"1-6e609020e0371257432797b4319c5829"}`),
					},
				},
			},
		},
		{
			name:    "all docs with simple string endkey",
			path:    "/_all_docs",
			options: kivik.Param("endkey", "foo"),
			db: newCustomDB(func(r *http.Request) (*http.Response, error) {
				if d := testy.DiffAsJSON([]byte(`"foo"`), []byte(r.URL.Query().Get("endkey"))); d != nil {
					t.Error(d)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Transfer-Encoding":  {"chunked"},
						"Date":               {"Sat, 01 Sep 2018 19:01:30 GMT"},
						"Server":             {"CouchDB/2.2.0 (Erlang OTP/19)"},
						"Content-Type":       {typeJSON},
						"Cache-Control":      {"must-revalidate"},
						"X-Couch-Request-ID": {"24fdb3fd86"},
						"X-Couch-Body-Time":  {"0"},
					},
					Body: io.NopCloser(strings.NewReader(`{"total_rows":1,"offset":null,"rows":[
{"id":"_design/_auth","key":"_design/_auth","value":{"rev":"1-6e609020e0371257432797b4319c5829"}}
]}`)),
				}, nil
			}),
			expected: queryResult{
				TotalRows: 1,
				UpdateSeq: "",
				Rows: []*driver.Row{
					{
						ID:    "_design/_auth",
						Key:   []byte(`"_design/_auth"`),
						Value: strings.NewReader(`{"rev":"1-6e609020e0371257432797b4319c5829"}`),
					},
				},
			},
		},
		{
			name:    "all docs with raw JSON endkey",
			path:    "/_all_docs",
			options: kivik.Param("endkey", json.RawMessage(`"foo"`)),
			db: newCustomDB(func(r *http.Request) (*http.Response, error) {
				if d := testy.DiffAsJSON([]byte(`"foo"`), []byte(r.URL.Query().Get("endkey"))); d != nil {
					t.Error(d)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Transfer-Encoding":  {"chunked"},
						"Date":               {"Sat, 01 Sep 2018 19:01:30 GMT"},
						"Server":             {"CouchDB/2.2.0 (Erlang OTP/19)"},
						"Content-Type":       {typeJSON},
						"Cache-Control":      {"must-revalidate"},
						"X-Couch-Request-ID": {"24fdb3fd86"},
						"X-Couch-Body-Time":  {"0"},
					},
					Body: io.NopCloser(strings.NewReader(`{"total_rows":1,"offset":null,"rows":[
{"id":"_design/_auth","key":"_design/_auth","value":{"rev":"1-6e609020e0371257432797b4319c5829"}}
]}`)),
				}, nil
			}),
			expected: queryResult{
				TotalRows: 1,
				UpdateSeq: "",
				Rows: []*driver.Row{
					{
						ID:    "_design/_auth",
						Key:   []byte(`"_design/_auth"`),
						Value: strings.NewReader(`{"rev":"1-6e609020e0371257432797b4319c5829"}`),
					},
				},
			},
		},
		{
			name:    "all docs with object keys",
			path:    "/_all_docs",
			options: kivik.Param("keys", []interface{}{"_design/_auth", "foo", []string{"bar", "baz"}}),
			db: newCustomDB(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPost {
					t.Errorf("Unexpected method: %s", r.Method)
				}
				defer r.Body.Close() // nolint: errcheck
				if d := testy.DiffAsJSON(map[string][]interface{}{"keys": {"_design/_auth", "foo", []string{"bar", "baz"}}}, r.Body); d != nil {
					t.Error(d)
				}
				if keys := r.URL.Query().Get("keys"); keys != "" {
					t.Error("query key 'keys' should be absent")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Transfer-Encoding":  {"chunked"},
						"Date":               {"Sat, 01 Sep 2018 19:01:30 GMT"},
						"Server":             {"CouchDB/2.2.0 (Erlang OTP/19)"},
						"Content-Type":       {typeJSON},
						"Cache-Control":      {"must-revalidate"},
						"X-Couch-Request-ID": {"24fdb3fd86"},
						"X-Couch-Body-Time":  {"0"},
					},
					Body: io.NopCloser(strings.NewReader(`{"total_rows":1,"offset":null,"rows":[
{"id":"_design/_auth","key":"_design/_auth","value":{"rev":"1-6e609020e0371257432797b4319c5829"}}
]}`)),
				}, nil
			}),
			expected: queryResult{
				TotalRows: 1,
				UpdateSeq: "",
				Rows: []*driver.Row{
					{
						ID:    "_design/_auth",
						Key:   []byte(`"_design/_auth"`),
						Value: strings.NewReader(`{"rev":"1-6e609020e0371257432797b4319c5829"}`),
					},
				},
			},
		},
		{
			name:    "all docs with docs",
			path:    "/_all_docs",
			options: kivik.Param("keys", []string{"_design/_auth", "foo"}),
			db: newCustomDB(func(r *http.Request) (*http.Response, error) {
				if r.Method != http.MethodPost {
					t.Errorf("Unexpected method: %s", r.Method)
				}
				defer r.Body.Close() // nolint: errcheck
				if d := testy.DiffAsJSON(map[string][]string{"keys": {"_design/_auth", "foo"}}, r.Body); d != nil {
					t.Error(d)
				}
				if keys := r.URL.Query().Get("keys"); keys != "" {
					t.Error("query key 'keys' should be absent")
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Transfer-Encoding":  {"chunked"},
						"Date":               {"Sat, 01 Sep 2018 19:01:30 GMT"},
						"Server":             {"CouchDB/2.2.0 (Erlang OTP/19)"},
						"Content-Type":       {typeJSON},
						"Cache-Control":      {"must-revalidate"},
						"X-Couch-Request-ID": {"24fdb3fd86"},
						"X-Couch-Body-Time":  {"0"},
					},
					Body: io.NopCloser(strings.NewReader(`{"total_rows":1,"offset":null,"rows":[
{"id":"foo","doc":{"_id":"foo"}}
]}`)),
				}, nil
			}),
			expected: queryResult{
				TotalRows: 1,
				UpdateSeq: "",
				Rows: []*driver.Row{
					{
						ID:  "foo",
						Doc: strings.NewReader(`{"_id":"foo"}`),
					},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			opts := test.options
			if opts == nil {
				opts = mock.NilOption
			}
			rows, err := test.db.rowsQuery(context.Background(), test.path, opts)
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if err != nil {
				return
			}
			result := queryResult{
				Rows: []*driver.Row{},
			}
			for {
				var row driver.Row
				if e := rows.Next(&row); e != nil {
					if e != io.EOF {
						result.Err = e.Error()
					}
					break
				}
				result.Rows = append(result.Rows, &row)
			}
			result.Offset = rows.Offset()
			result.TotalRows = rows.TotalRows()
			result.UpdateSeq = rows.UpdateSeq()
			if warner, ok := rows.(driver.RowsWarner); ok {
				result.Warning = warner.Warning()
			} else {
				t.Errorf("RowsWarner interface not satisfied!!?")
			}

			if d := queryResultDiff(test.expected, result); d != "" {
				t.Error(d)
			}
		})
	}
}

func TestSecurity(t *testing.T) {
	tests := []struct {
		name     string
		db       *db
		expected *driver.Security
		status   int
		err      string
	}{
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/testdb/_security"?: net error`,
		},
		{
			name: "1.6.1 empty",
			db: newTestDB(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":           {"Thu, 26 Oct 2017 14:28:14 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"3"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: io.NopCloser(strings.NewReader("{}")),
			}, nil),
			expected: &driver.Security{},
		},
		{
			name: "1.6.1 non-empty",
			db: newTestDB(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":           {"Thu, 26 Oct 2017 14:28:14 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"65"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: io.NopCloser(strings.NewReader(`{"admins":{},"members":{"names":["32dgsme3cmi6pddghslq5egiye"]}}`)),
			}, nil),
			expected: &driver.Security{
				Members: driver.Members{
					Names: []string{"32dgsme3cmi6pddghslq5egiye"},
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Security(context.Background())
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestSetSecurity(t *testing.T) {
	type tt struct {
		db       *db
		security *driver.Security
		status   int
		err      string
	}
	tests := testy.NewTable()

	tests.Add("network error", tt{
		db:     newTestDB(nil, errors.New("net error")),
		status: http.StatusBadGateway,
		err:    `Put "?http://example.com/testdb/_security"?: net error`,
	})
	tests.Add("1.6.1", func(t *testing.T) interface{} {
		return tt{
			security: &driver.Security{
				Admins: driver.Members{
					Names: []string{"bob"},
				},
				Members: driver.Members{
					Roles: []string{"users"},
				},
			},
			db: newCustomDB(func(req *http.Request) (*http.Response, error) {
				defer req.Body.Close() // nolint: errcheck
				if ct, _, _ := mime.ParseMediaType(req.Header.Get("Content-Type")); ct != typeJSON {
					return nil, fmt.Errorf("Expected Content-Type: application/json, got %s", ct)
				}
				var body interface{}
				if err := json.NewDecoder(req.Body).Decode(&body); err != nil {
					return nil, err
				}
				expected := map[string]interface{}{
					"admins": map[string]interface{}{
						"names": []string{"bob"},
					},
					"members": map[string]interface{}{
						"roles": []string{"users"},
					},
				}
				if d := testy.DiffAsJSON(expected, body); d != nil {
					t.Error(d)
				}
				return &http.Response{
					StatusCode: 200,
					Header: http.Header{
						"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
						"Date":           {"Thu, 26 Oct 2017 15:06:21 GMT"},
						"Content-Type":   {"text/plain; charset=utf-8"},
						"Content-Length": {"12"},
						"Cache-Control":  {"must-revalidate"},
					},
					Body: io.NopCloser(strings.NewReader(`{"ok":true}`)),
				}, nil
			}),
		}
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		err := tt.db.SetSecurity(context.Background(), tt.security)
		if d := internal.StatusErrorDiffRE(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
	})
}

func TestGetRev(t *testing.T) {
	tests := []struct {
		name   string
		db     *db
		id     string
		rev    string
		status int
		err    string
	}{
		{
			name:   "no doc id",
			status: http.StatusBadRequest,
			err:    "kivik: docID required",
		},
		{
			name:   "network error",
			id:     "foo",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Head "?http://example.com/testdb/foo"?: net error`,
		},
		{
			name: "1.6.1",
			id:   "foo",
			db: newTestDB(&http.Response{
				StatusCode: 200,
				Request: &http.Request{
					Method: "HEAD",
				},
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"ETag":           {`"1-4c6114c65e295552ab1019e2b046b10e"`},
					"Date":           {"Thu, 26 Oct 2017 15:21:15 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"70"},
					"Cache-Control":  {"must-revalidate"},
				},
				ContentLength: 70,
				Body:          io.NopCloser(strings.NewReader("")),
			}, nil),
			rev: "1-4c6114c65e295552ab1019e2b046b10e",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rev, err := test.db.GetRev(context.Background(), test.id, mock.NilOption)
			if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if rev != test.rev {
				t.Errorf("Got rev %s, expected %s", rev, test.rev)
			}
		})
	}
}

func TestCopy(t *testing.T) {
	type tt struct {
		target, source string
		options        kivik.Option
		db             *db
		rev            string
		status         int
		err            string
	}

	tests := testy.NewTable()
	tests.Add("missing source", tt{
		status: http.StatusBadRequest,
		err:    "kivik: sourceID required",
	})
	tests.Add("missing target", tt{
		source: "foo",
		status: http.StatusBadRequest,
		err:    "kivik: targetID required",
	})
	tests.Add("network error", tt{
		source: "foo",
		target: "bar",
		db:     newTestDB(nil, errors.New("net error")),
		status: http.StatusBadGateway,
		err:    "(Copy http://example.com/testdb/foo: )?net error",
	})
	tests.Add("invalid options", tt{
		db:      &db{},
		source:  "foo",
		target:  "bar",
		options: kivik.Param("foo", make(chan int)),
		status:  http.StatusBadRequest,
		err:     "kivik: invalid type chan int for options",
	})
	tests.Add("create 1.6.1", tt{
		source: "foo",
		target: "bar",
		db: newCustomDB(func(req *http.Request) (*http.Response, error) {
			if req.Header.Get("Destination") != "bar" {
				return nil, errors.New("Unexpected destination")
			}
			return &http.Response{
				StatusCode: 201,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Location":       {"http://example.com/foo/bar"},
					"ETag":           {`"1-f81c8a795b0c6f9e9f699f64c6b82256"`},
					"Date":           {"Thu, 26 Oct 2017 15:45:57 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"66"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: Body(`{"ok":true,"id":"bar","rev":"1-f81c8a795b0c6f9e9f699f64c6b82256"}`),
			}, nil
		}),
		rev: "1-f81c8a795b0c6f9e9f699f64c6b82256",
	})
	tests.Add("full commit 1.6.1", tt{
		source:  "foo",
		target:  "bar",
		options: OptionFullCommit(),
		db: newCustomDB(func(req *http.Request) (*http.Response, error) {
			if dest := req.Header.Get("Destination"); dest != "bar" {
				return nil, fmt.Errorf("Unexpected destination: %s", dest)
			}
			if fc := req.Header.Get("X-Couch-Full-Commit"); fc != "true" {
				return nil, fmt.Errorf("X-Couch-Full-Commit: %s", fc)
			}
			return &http.Response{
				StatusCode: 201,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Location":       {"http://example.com/foo/bar"},
					"ETag":           {`"1-f81c8a795b0c6f9e9f699f64c6b82256"`},
					"Date":           {"Thu, 26 Oct 2017 15:45:57 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"66"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: Body(`{"ok":true,"id":"bar","rev":"1-f81c8a795b0c6f9e9f699f64c6b82256"}`),
			}, nil
		}),
		rev: "1-f81c8a795b0c6f9e9f699f64c6b82256",
	})
	tests.Add("target rev", tt{
		source:  "foo",
		target:  "bar?rev=1-xxx",
		options: OptionFullCommit(),
		db: newCustomDB(func(req *http.Request) (*http.Response, error) {
			if dest := req.Header.Get("Destination"); dest != "bar?rev=1-xxx" {
				return nil, fmt.Errorf("Unexpected destination: %s", dest)
			}
			if fc := req.Header.Get("X-Couch-Full-Commit"); fc != "true" {
				return nil, fmt.Errorf("X-Couch-Full-Commit: %s", fc)
			}
			return &http.Response{
				StatusCode: 201,
				Header: http.Header{
					"Server":         {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Location":       {"http://example.com/foo/bar"},
					"ETag":           {`"2-yyy"`},
					"Date":           {"Thu, 26 Oct 2017 15:45:57 GMT"},
					"Content-Type":   {"text/plain; charset=utf-8"},
					"Content-Length": {"66"},
					"Cache-Control":  {"must-revalidate"},
				},
				Body: Body(`{"ok":true,"id":"bar","rev":"2-yyy"}`),
			}, nil
		}),
		rev: "2-yyy",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		opts := tt.options
		if opts == nil {
			opts = mock.NilOption
		}
		rev, err := tt.db.Copy(context.Background(), tt.target, tt.source, opts)
		if d := internal.StatusErrorDiffRE(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if rev != tt.rev {
			t.Errorf("Got %s, expected %s", rev, tt.rev)
		}
	})
}

func TestMultipartAttachmentsNext(t *testing.T) {
	tests := []struct {
		name     string
		atts     *multipartAttachments
		content  string
		expected *driver.Attachment
		status   int
		err      string
	}{
		{
			name: "done reading",
			atts: &multipartAttachments{
				mpReader: func() *multipart.Reader {
					r := multipart.NewReader(strings.NewReader("--xxx\r\n\r\n--xxx--"), "xxx")
					_, _ = r.NextPart()
					return r
				}(),
			},
			status: 500,
			err:    io.EOF.Error(),
		},
		{
			name: "malformed message",
			atts: &multipartAttachments{
				mpReader: func() *multipart.Reader {
					r := multipart.NewReader(strings.NewReader("oink"), "xxx")
					_, _ = r.NextPart()
					return r
				}(),
			},
			status: http.StatusBadGateway,
			err:    "multipart: NextPart: EOF",
		},
		{
			name: "malformed Content-Disposition",
			atts: &multipartAttachments{
				mpReader: multipart.NewReader(strings.NewReader(`--xxx
Content-Type: text/plain

--xxx--`), "xxx"),
			},
			status: http.StatusBadGateway,
			err:    "Content-Disposition: mime: no media type",
		},
		{
			name: "malformed Content-Type",
			atts: &multipartAttachments{
				meta: map[string]attMeta{
					"foo.txt": {Follows: true},
				},
				mpReader: multipart.NewReader(strings.NewReader(`--xxx
Content-Type: text/plain; =foo
Content-Disposition: attachment; filename="foo.txt"

--xxx--`), "xxx"),
			},
			status: http.StatusBadGateway,
			err:    "mime: invalid media parameter",
		},
		{
			name: "file not in manifest",
			atts: &multipartAttachments{
				mpReader: multipart.NewReader(strings.NewReader(`--xxx
Content-Type: text/plain; charset=foobar
Content-Disposition: attachment; filename="foo.txt"

test content
--xxx--`), "xxx"),
			},
			status: http.StatusBadGateway,
			err:    "File 'foo.txt' not in manifest",
		},
		{
			name: "invalid content-disposition",
			atts: &multipartAttachments{
				mpReader: multipart.NewReader(strings.NewReader(`--xxx
Content-Type: text/plain
Content-Disposition: oink

--xxx--`), "xxx"),
			},
			status: http.StatusBadGateway,
			err:    "Unexpected Content-Disposition: oink",
		},
		{
			name: "success",
			atts: &multipartAttachments{
				meta: map[string]attMeta{
					"foo.txt": {Follows: true},
				},
				mpReader: multipart.NewReader(strings.NewReader(`--xxx
Content-Type: text/plain; charset=foobar
Content-Disposition: attachment; filename="foo.txt"

test content
--xxx--`), "xxx"),
			},
			content: "test content",
			expected: &driver.Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Size:        -1,
			},
		},
		{
			name: "success, no Content-Type header, & Content-Length header",
			atts: &multipartAttachments{
				meta: map[string]attMeta{
					"foo.txt": {
						Follows:     true,
						ContentType: "text/plain",
					},
				},
				mpReader: multipart.NewReader(strings.NewReader(`--xxx
Content-Disposition: attachment; filename="foo.txt"
Content-Length: 123

test content
--xxx--`), "xxx"),
			},
			content: "test content",
			expected: &driver.Attachment{
				Filename:    "foo.txt",
				ContentType: "text/plain",
				Size:        123,
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := new(driver.Attachment)
			err := test.atts.Next(result)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if err != nil {
				return
			}
			content, err := io.ReadAll(result.Content)
			if err != nil {
				t.Fatal(err)
			}
			if d := testy.DiffText(test.content, string(content)); d != nil {
				t.Errorf("Unexpected content:\n%s", d)
			}
			result.Content = nil // Determinism
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestMultipartAttachmentsClose(t *testing.T) {
	const wantErr = "some error"
	atts := &multipartAttachments{
		content: &mockReadCloser{
			CloseFunc: func() error {
				return errors.New(wantErr)
			},
		},
	}

	if err := atts.Close(); !testy.ErrorMatches(wantErr, err) {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestPurge(t *testing.T) {
	expectedDocMap := map[string][]string{
		"foo": {"1-abc", "2-def"},
		"bar": {"3-ghi"},
	}
	tests := []struct {
		name   string
		db     *db
		docMap map[string][]string

		expected *driver.PurgeResult
		err      string
		status   int
	}{
		{
			name: "1.7.1, nothing deleted",
			db: newCustomDB(func(r *http.Request) (*http.Response, error) {
				if r.Method != "POST" {
					return nil, fmt.Errorf("Unexpected method: %s", r.Method)
				}
				if r.URL.Path != "/testdb/_purge" {
					return nil, fmt.Errorf("Unexpected path: %s", r.URL.Path)
				}
				if ct := r.Header.Get("Content-Type"); ct != typeJSON {
					return nil, fmt.Errorf("Unexpected Content-Type: %s", ct)
				}
				defer r.Body.Close() // nolint: errcheck
				var result interface{}
				if err := json.NewDecoder(r.Body).Decode(&result); err != nil {
					return nil, err
				}
				if d := testy.DiffAsJSON(expectedDocMap, result); d != nil {
					return nil, fmt.Errorf("Unexpected payload:\n%s", d)
				}
				return &http.Response{
					StatusCode: http.StatusOK,
					Header: http.Header{
						"Server":         []string{"CouchDB/1.7.1 (Erlang OTP/17)"},
						"Date":           []string{"Thu, 06 Sep 2018 16:55:26 GMT"},
						"Content-Type":   []string{"text/plain; charset=utf-8"},
						"Content-Length": []string{"28"},
						"Cache-Control":  []string{"must-revalidate"},
					},
					Body: io.NopCloser(strings.NewReader(`{"purge_seq":3,"purged":{}}`)),
				}, nil
			}),
			docMap:   expectedDocMap,
			expected: &driver.PurgeResult{Seq: 3, Purged: map[string][]string{}},
		},
		{
			name: "1.7.1, all deleted",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusOK,
				Header: http.Header{
					"Server":         []string{"CouchDB/1.7.1 (Erlang OTP/17)"},
					"Date":           []string{"Thu, 06 Sep 2018 16:55:26 GMT"},
					"Content-Type":   []string{"text/plain; charset=utf-8"},
					"Content-Length": []string{"168"},
					"Cache-Control":  []string{"must-revalidate"},
				},
				Body: io.NopCloser(strings.NewReader(`{"purge_seq":5,"purged":{"foo":["1-abc","2-def"],"bar":["3-ghi"]}}`)),
			}, nil),
			docMap:   expectedDocMap,
			expected: &driver.PurgeResult{Seq: 5, Purged: expectedDocMap},
		},
		{
			name: "2.2.0, not supported",
			db: newTestDB(&http.Response{
				StatusCode:    501,
				ContentLength: 75,
				Header: http.Header{
					"Server":              []string{"CouchDB/2.2.0 (Erlang OTP/19)"},
					"Date":                []string{"Thu, 06 Sep 2018 16:55:26 GMT"},
					"Content-Type":        []string{typeJSON},
					"Content-Length":      []string{"75"},
					"Cache-Control":       []string{"must-revalidate"},
					"X-Couch-Request-ID":  []string{"03e91291c8"},
					"X-CouchDB-Body-Time": []string{"0"},
				},
				Body: io.NopCloser(strings.NewReader(`{"error":"not_implemented","reason":"this feature is not yet implemented"}`)),
			}, nil),
			docMap: expectedDocMap,
			err:    "Not Implemented: this feature is not yet implemented",
			status: http.StatusNotImplemented,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.db.Purge(context.Background(), test.docMap)
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if err != nil {
				return
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestMultipartAttachments(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		atts     *kivik.Attachments
		expected string
		size     int64
		err      string
	}{
		{
			name:  "no attachments",
			input: `{"foo":"bar","baz":"qux"}`,
			atts:  &kivik.Attachments{},
			expected: `
--%[1]s
Content-Type: application/json

{"foo":"bar","baz":"qux"}
--%[1]s--
`,
			size: 191,
		},
		{
			name:  "simple",
			input: `{"_attachments":{}}`,
			atts: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
			},
			expected: `
--%[1]s
Content-Type: application/json

{"_attachments":{"foo.txt":{"content_type":"text/plain","length":13,"follows":true}}
}
--%[1]s

test content

--%[1]s--
`,
			size: 333,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			in := io.NopCloser(strings.NewReader(test.input))
			boundary, size, body, err := newMultipartAttachments(in, test.atts)
			if !testy.ErrorMatches(test.err, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			if test.size != size {
				t.Errorf("Unexpected size: %d (want %d)", size, test.size)
			}
			result, _ := io.ReadAll(body)
			expected := fmt.Sprintf(test.expected, boundary)
			expected = strings.TrimPrefix(expected, "\n")
			result = bytes.ReplaceAll(result, []byte("\r\n"), []byte("\n"))
			if d := testy.DiffText(expected, string(result)); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestAttachmentStubs(t *testing.T) {
	tests := []struct {
		name     string
		atts     *kivik.Attachments
		expected map[string]*stub
	}{
		{
			name: "simple",
			atts: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
			},
			expected: map[string]*stub{
				"foo.txt": {
					ContentType: "text/plain",
					Size:        13,
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, _ := attachmentStubs(test.atts)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestInterfaceToAttachments(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		output   interface{}
		expected *kivik.Attachments
		ok       bool
	}{
		{
			name:     "non-attachment input",
			input:    "foo",
			output:   "foo",
			expected: nil,
			ok:       false,
		},
		{
			name: "pointer input",
			input: &kivik.Attachments{
				"foo.txt": nil,
			},
			output: new(kivik.Attachments),
			expected: &kivik.Attachments{
				"foo.txt": nil,
			},
			ok: true,
		},
		{
			name: "non-pointer input",
			input: kivik.Attachments{
				"foo.txt": nil,
			},
			output: kivik.Attachments{},
			expected: &kivik.Attachments{
				"foo.txt": nil,
			},
			ok: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, ok := interfaceToAttachments(test.input)
			if ok != test.ok {
				t.Errorf("Unexpected OK result: %v", result)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Errorf("Unexpected result:\n%s\n", d)
			}
			if d := testy.DiffInterface(test.output, test.input); d != nil {
				t.Errorf("Input not properly modified:\n%s\n", d)
			}
		})
	}
}

func TestStubMarshalJSON(t *testing.T) {
	att := &stub{
		ContentType: "text/plain",
		Size:        123,
	}
	expected := `{"content_type":"text/plain","length":123,"follows":true}`
	result, err := json.Marshal(att)
	if !testy.ErrorMatches("", err) {
		t.Errorf("Unexpected error: %s", err)
	}
	if d := testy.DiffJSON([]byte(expected), result); d != nil {
		t.Error(d)
	}
}

func Test_attachmentSize(t *testing.T) {
	type tst struct {
		att      *kivik.Attachment
		expected *kivik.Attachment
		err      string
	}
	tests := testy.NewTable()
	tests.Add("size already set", tst{
		att:      &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("text"), Size: 4},
		expected: &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("text"), Size: 4},
	})
	tests.Add("bytes buffer", tst{
		att:      &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("text")},
		expected: &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("text"), Size: 5},
	})
	tests.Run(t, func(t *testing.T, test tst) {
		err := attachmentSize(test.att)
		if !testy.ErrorMatches(test.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		body, err := io.ReadAll(test.att.Content)
		if err != nil {
			t.Fatal(err)
		}
		expBody, err := io.ReadAll(test.expected.Content)
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffText(expBody, body); d != nil {
			t.Errorf("Content differs:\n%s\n", d)
		}
		test.att.Content = nil
		test.expected.Content = nil
		if d := testy.DiffInterface(test.expected, test.att); d != nil {
			t.Error(d)
		}
	})
}

type lenReader interface {
	io.Reader
	lener
}

type myReader struct {
	lenReader
}

var _ interface {
	io.Closer
	lenReader
} = &myReader{}

func (r *myReader) Close() error { return nil }

func Test_readerSize(t *testing.T) {
	type tst struct {
		in   io.ReadCloser
		size int64
		body string
		err  string
	}
	tests := testy.NewTable()
	tests.Add("*bytes.Buffer", tst{
		in:   &myReader{bytes.NewBuffer([]byte("foo bar"))},
		size: 7,
		body: "foo bar",
	})
	tests.Add("bytes.NewReader", tst{
		in:   &myReader{bytes.NewReader([]byte("foo bar"))},
		size: 7,
		body: "foo bar",
	})
	tests.Add("strings.NewReader", tst{
		in:   &myReader{strings.NewReader("foo bar")},
		size: 7,
		body: "foo bar",
	})
	tests.Add("file", func(t *testing.T) interface{} {
		f, err := os.CreateTemp("", "file-reader-*")
		if err != nil {
			t.Fatal(err)
		}
		tests.Cleanup(func() {
			_ = os.Remove(f.Name())
		})
		if _, err := f.Write([]byte("foo bar")); err != nil {
			t.Fatal(err)
		}
		if _, err := f.Seek(0, 0); err != nil {
			t.Fatal(err)
		}
		return tst{
			in:   f,
			size: 7,
			body: "foo bar",
		}
	})
	tests.Add("nop closer", tst{
		in:   io.NopCloser(strings.NewReader("foo bar")),
		size: 7,
		body: "foo bar",
	})
	tests.Add("seeker", tst{
		in:   &seeker{strings.NewReader("asdf asdf")},
		size: 9,
		body: "asdf asdf",
	})
	tests.Run(t, func(t *testing.T, test tst) {
		size, r, err := readerSize(test.in)
		if !testy.ErrorMatches(test.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		body, err := io.ReadAll(r)
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffText(test.body, body); d != nil {
			t.Errorf("Unexpected body content:\n%s\n", d)
		}
		if size != test.size {
			t.Errorf("Unexpected size: %d\n", size)
		}
	})
}

type seeker struct {
	r *strings.Reader
}

func (s *seeker) Read(b []byte) (int, error) {
	return s.r.Read(b)
}

func (s *seeker) Seek(offset int64, whence int) (int64, error) {
	return s.r.Seek(offset, whence)
}

func (s *seeker) Close() error { return nil }

func TestNewAttachment(t *testing.T) {
	type tst struct {
		content    io.Reader
		size       []int64
		expected   *kivik.Attachment
		expContent string
		err        string
	}
	tests := testy.NewTable()
	tests.Add("size provided", tst{
		content: strings.NewReader("xxx"),
		size:    []int64{99},
		expected: &kivik.Attachment{
			Filename:    "foo.txt",
			ContentType: "text/plain",
			Size:        99,
		},
		expContent: "xxx",
	})
	tests.Add("strings.NewReader", tst{
		content: strings.NewReader("xxx"),
		expected: &kivik.Attachment{
			Filename:    "foo.txt",
			ContentType: "text/plain",
			Size:        3,
		},
		expContent: "xxx",
	})
	tests.Run(t, func(t *testing.T, test tst) {
		result, err := NewAttachment("foo.txt", "text/plain", test.content, test.size...)
		if !testy.ErrorMatches(test.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		content, err := io.ReadAll(result.Content)
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffText(test.expContent, content); d != nil {
			t.Errorf("Unexpected content:\n%s\n", d)
		}
		result.Content = nil
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestCopyWithAttachmentStubs(t *testing.T) {
	type tst struct {
		input    io.Reader
		w        io.Writer
		expected string
		atts     map[string]*stub
		status   int
		err      string
	}
	tests := testy.NewTable()
	tests.Add("no attachments", tst{
		input:    strings.NewReader("{}"),
		expected: "{}",
	})
	tests.Add("Unexpected delim", tst{
		input:  strings.NewReader("[]"),
		status: http.StatusBadRequest,
		err:    `^expected '{', found '\['$`,
	})
	tests.Add("read error", tst{
		input:  testy.ErrorReader("", errors.New("read error")),
		status: http.StatusInternalServerError,
		err:    "^read error$",
	})
	tests.Add("write error", tst{
		input:  strings.NewReader("{}"),
		w:      testy.ErrorWriter(0, errors.New("write error")),
		status: http.StatusInternalServerError,
		err:    "^write error$",
	})
	tests.Add("decode error", tst{
		input:  strings.NewReader("{}}"),
		status: http.StatusBadRequest,
		err:    "^invalid character '}' +looking for beginning of value$",
	})
	tests.Add("one attachment", tst{
		input: strings.NewReader(`{"_attachments":{}}`),
		atts: map[string]*stub{
			"foo.txt": {
				ContentType: "text/plain",
				Size:        3,
			},
		},
		expected: `{"_attachments":{"foo.txt":{"content_type":"text/plain","length":3,"follows":true}}
}`,
	})

	tests.Run(t, func(t *testing.T, test tst) {
		w := test.w
		if w == nil {
			w = &bytes.Buffer{}
		}
		err := copyWithAttachmentStubs(w, test.input, test.atts)
		if d := internal.StatusErrorDiffRE(test.err, test.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		if d := testy.DiffText(test.expected, w.(*bytes.Buffer).String()); d != nil {
			t.Error(d)
		}
	})
}

func TestRevsDiff(t *testing.T) {
	type tt struct {
		db     *db
		revMap map[string][]string
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("net error", tt{
		db:     newTestDB(nil, errors.New("net error")),
		status: http.StatusBadGateway,
		err:    `Post "?http://example.com/testdb/_revs_diff"?: net error`,
	})
	tests.Add("success", tt{
		db: newCustomDB(func(r *http.Request) (*http.Response, error) {
			expectedBody := json.RawMessage(`{
				"190f721ca3411be7aa9477db5f948bbb": [
					"3-bb72a7682290f94a985f7afac8b27137",
					"4-10265e5a26d807a3cfa459cf1a82ef2e",
					"5-067a00dff5e02add41819138abb3284d"
				]
			}`)
			defer r.Body.Close() // nolint: errcheck
			if d := testy.DiffAsJSON(expectedBody, r.Body); d != nil {
				return nil, fmt.Errorf("Unexpected payload: %s", d)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body: io.NopCloser(strings.NewReader(`{
					"190f721ca3411be7aa9477db5f948bbb": {
						"missing": [
							"3-bb72a7682290f94a985f7afac8b27137",
							"5-067a00dff5e02add41819138abb3284d"
						],
						"possible_ancestors": [
							"4-10265e5a26d807a3cfa459cf1a82ef2e"
						]
					},
					"foo": {
						"missing": ["1-xxx"]
					}
				}`)),
			}, nil
		}),
		revMap: map[string][]string{
			"190f721ca3411be7aa9477db5f948bbb": {
				"3-bb72a7682290f94a985f7afac8b27137",
				"4-10265e5a26d807a3cfa459cf1a82ef2e",
				"5-067a00dff5e02add41819138abb3284d",
			},
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		ctx, cancel := context.WithTimeout(context.TODO(), 2*time.Second)
		defer cancel()
		rows, err := tt.db.RevsDiff(ctx, tt.revMap)
		if d := internal.StatusErrorDiffRE(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		results := make(map[string]interface{})
		drow := new(driver.Row)
		for {
			if err := rows.Next(drow); err != nil {
				if err == io.EOF {
					break
				}
				t.Fatal(err)
			}
			var row interface{}
			if err := json.NewDecoder(drow.Value).Decode(&row); err != nil {
				t.Fatal(err)
			}
			results[drow.ID] = row
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), results); d != nil {
			t.Error(d)
		}
	})
}
