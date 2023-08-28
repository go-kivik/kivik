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
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
)

func TestChanges_metadata(t *testing.T) {
	db := newTestDB(&http.Response{
		StatusCode: 200,
		Header:     http.Header{},
		Body: Body(`{"results":[
			{"seq":"1-g1AAAABteJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5kTEXKMBuZmKebGSehK4Yh_Y8FiDJ0ACk_oNMSWTIAgDjASHc","id":"56d164e9566e12cb9dff87d455000f3d","changes":[{"rev":"1-967a00dff5e02add41819138abb3284d"}]},
			{"seq":"2-g1AAAACLeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5kTEXKMBuZmKebGSehK4Yh_Y8FiDJ0ACk_qOYYm5qYGBklIquJwsAO5gqIA","id":"56d164e9566e12cb9dff87d455001b58","changes":[{"rev":"1-967a00dff5e02add41819138abb3284d"}]},
			{"seq":"3-g1AAAACLeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5kSkXKMBuZmKebGSehK4Yh_Y8FiDJ0ACk_kNNYQSbYm5qYGBklIquJwsAO_wqIQ","id":"56d164e9566e12cb9dff87d455002462","changes":[{"rev":"1-967a00dff5e02add41819138abb3284d"}]},
			{"seq":"4-g1AAAACLeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5kSkXKMBuZmKebGSehK4Yh_Y8FiDJ0ACk_qOYYm5qYGBklIquJwsAPBoqIg","id":"56d164e9566e12cb9dff87d455004150","changes":[{"rev":"1-967a00dff5e02add41819138abb3284d"}]},
			{"seq":"5-g1AAAACLeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5kTkXKMBuZmKebGSehK4Yh_Y8FiDJ0ACk_kNNYQKbYm5qYGBklIquJwsAPH4qIw","id":"56d164e9566e12cb9dff87d455003421","changes":[{"rev":"1-967a00dff5e02add41819138abb3284d"}]}
			],
			"last_seq":"5-g1AAAACLeJzLYWBgYMpgTmHgzcvPy09JdcjLz8gvLskBCScyJNX___8_K4M5kTkXKMBuZmKebGSehK4Yh_Y8FiDJ0ACk_kNNYQKbYm5qYGBklIquJwsAPH4qIw","pending":10}
		`),
	}, nil)

	changes, err := db.Changes(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	ch := &driver.Change{}
	for {
		if changes.Next(ch) != nil {
			break
		}
	}
	want := int64(10)
	if got := changes.Pending(); want != got {
		t.Errorf("want: %d, got: %d", want, got)
	}
}

func TestChanges(t *testing.T) {
	tests := []struct {
		name    string
		options map[string]interface{}
		db      *db
		status  int
		err     string
		etag    string
	}{
		{
			name:    "invalid options",
			options: map[string]interface{}{"foo": make(chan int)},
			status:  http.StatusBadRequest,
			err:     "kivik: invalid type chan int for options",
		},
		{
			name:    "eventsource",
			options: map[string]interface{}{"feed": "eventsource"},
			status:  http.StatusBadRequest,
			err:     "kivik: eventsource feed not supported, use 'continuous'",
		},
		{
			name:   "network error",
			db:     newTestDB(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/testdb/_changes"?: net error`,
		},
		{
			name:    "continuous",
			db:      newTestDB(nil, errors.New("net error")),
			options: map[string]interface{}{"feed": "continuous"},
			status:  http.StatusBadGateway,
			err:     `Get "?http://example.com/testdb/_changes\?feed=continuous"?: net error`,
		},
		{
			name: "error response",
			db: newTestDB(&http.Response{
				StatusCode: http.StatusBadRequest,
				Body:       Body(""),
			}, nil),
			status: http.StatusBadRequest,
			err:    "Bad Request",
		},
		{
			name: "success 1.6.1",
			db: newTestDB(&http.Response{
				StatusCode: 200,
				Header: http.Header{
					"Transfer-Encoding": {"chunked"},
					"Server":            {"CouchDB/1.6.1 (Erlang OTP/17)"},
					"Date":              {"Fri, 27 Oct 2017 14:43:57 GMT"},
					"Content-Type":      {"text/plain; charset=utf-8"},
					"Cache-Control":     {"must-revalidate"},
					"ETag":              {`"etag-foo"`},
				},
				Body: Body(`{"seq":3,"id":"43734cf3ce6d5a37050c050bb600006b","changes":[{"rev":"2-185ccf92154a9f24a4f4fd12233bf463"}],"deleted":true}
                    `),
			}, nil),
			etag: "etag-foo",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			ch, err := test.db.Changes(context.Background(), test.options)
			if ch != nil {
				defer ch.Close() // nolint:errcheck
			}
			testy.StatusErrorRE(t, test.err, test.status, err)
			if etag := ch.ETag(); etag != test.etag {
				t.Errorf("Unexpected ETag: %s", etag)
			}
		})
	}
}

func TestChangesNext(t *testing.T) {
	tests := []struct {
		name     string
		changes  *changesRows
		status   int
		err      string
		expected *driver.Change
	}{
		{
			name:    "invalid json",
			changes: newChangesRows(context.TODO(), "", Body("invalid json"), ""),
			status:  http.StatusBadGateway,
			err:     "invalid character 'i' looking for beginning of value",
		},
		{
			name: "success",
			changes: newChangesRows(context.TODO(), "", Body(`{"seq":3,"id":"43734cf3ce6d5a37050c050bb600006b","changes":[{"rev":"2-185ccf92154a9f24a4f4fd12233bf463"}],"deleted":true}
                `), ""),
			expected: &driver.Change{
				ID:      "43734cf3ce6d5a37050c050bb600006b",
				Seq:     "3",
				Deleted: true,
				Changes: []string{"2-185ccf92154a9f24a4f4fd12233bf463"},
			},
		},
		{
			name:    "read error",
			changes: newChangesRows(context.TODO(), "", io.NopCloser(testy.ErrorReader("", errors.New("read error"))), ""),
			status:  http.StatusBadGateway,
			err:     "read error",
		},
		{
			name:     "end of input",
			changes:  newChangesRows(context.TODO(), "", Body(``), ""),
			expected: &driver.Change{},
			status:   http.StatusInternalServerError,
			err:      "EOF",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			row := new(driver.Change)
			err := test.changes.Next(row)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, row); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestChangesClose(t *testing.T) {
	t.Run("normal", func(t *testing.T) {
		body := &closeTracker{ReadCloser: Body("foo")}
		feed := newChangesRows(context.TODO(), "", body, "")
		_ = feed.Close()
		if !body.closed {
			t.Errorf("Failed to close")
		}
	})

	t.Run("next in progress", func(t *testing.T) {
		body := &closeTracker{ReadCloser: io.NopCloser(testy.NeverReader())}
		feed := newChangesRows(context.TODO(), "", body, "")
		row := new(driver.Change)
		done := make(chan struct{})
		go func() {
			_ = feed.Next(row)
			close(done)
		}()
		time.Sleep(50 * time.Millisecond)
		_ = feed.Close()
		<-done
		if !body.closed {
			t.Errorf("Failed to close")
		}
	})
}
