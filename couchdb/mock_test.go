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
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/int/mock"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

const isGopherJS = runtime.GOOS == "js" || runtime.GOARCH == "js"

type customTransport func(*http.Request) (*http.Response, error)

var _ http.RoundTripper = customTransport(nil)

func (t customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t(req)
}

func newTestDB(response *http.Response, err error) *db {
	return &db{
		dbName: "testdb",
		client: newTestClient(response, err),
	}
}

func newCustomDB(fn func(*http.Request) (*http.Response, error)) *db {
	return &db{
		dbName: "testdb",
		client: newCustomClient(fn),
	}
}

func newTestClient(response *http.Response, err error) *client {
	return newCustomClient(func(req *http.Request) (*http.Response, error) {
		if e := consume(req.Body); e != nil {
			return nil, e
		}
		if err != nil {
			return nil, err
		}
		response := response
		response.Request = req
		return response, nil
	})
}

func newTestClientWithRawPath(path string, response *http.Response, err error) *client {
	return newCustomClient(func(req *http.Request) (*http.Response, error) {
		if e := consume(req.Body); e != nil {
			return nil, e
		}
		if err != nil {
			return nil, err
		}
		if req.URL.RawPath != path {
			return nil, fmt.Errorf("expected path %s, got %s", path, req.URL.Path)
		}
		response := response
		response.Request = req
		return response, nil
	})
}

func newCustomClient(fn func(*http.Request) (*http.Response, error)) *client {
	chttpClient, _ := chttp.New(&http.Client{}, "http://example.com/", OptionNoRequestCompression())
	chttpClient.Client.Transport = customTransport(fn)
	return &client{
		Client: chttpClient,
	}
}

func Body(str string) io.ReadCloser {
	if !strings.HasSuffix(str, "\n") {
		str += "\n"
	}
	return io.NopCloser(strings.NewReader(str))
}

func parseTime(t *testing.T, str string) time.Time {
	t.Helper()
	ts, err := time.Parse(time.RFC3339, str)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}

// consume consumes and closes r or does nothing if it is nil.
func consume(r io.ReadCloser) error {
	if r == nil {
		return nil
	}
	defer r.Close() // nolint: errcheck
	_, e := io.ReadAll(r)
	return e
}

type mockReadCloser struct {
	ReadFunc  func([]byte) (int, error)
	CloseFunc func() error
}

var _ io.ReadCloser = &mockReadCloser{}

func (rc *mockReadCloser) Read(p []byte) (int, error) {
	return rc.ReadFunc(p)
}

func (rc *mockReadCloser) Close() error {
	return rc.CloseFunc()
}

func realDB(t *testing.T) *db {
	t.Helper()
	if isGopherJS {
		t.Skip("Network tests skipped for GopherJS")
	}
	db, err := realDBConnect(t)
	if err != nil {
		if errors.Is(err, &url.Error{}) {
			t.Skip("Cannot connect to CouchDB")
		}
		if strings.HasSuffix(err.Error(), "connect: connection refused") {
			t.Skip("Cannot connect to CouchDB")
		}
		t.Fatal(err)
	}
	return db
}

func realDBConnect(t *testing.T) (*db, error) {
	t.Helper()
	driver := &couch{}
	c, err := driver.NewClient(kt.DSN3(t), OptionNoRequestCompression())
	if err != nil {
		return nil, err
	}
	dbname := kt.TestDBName(t)

	err = c.CreateDB(context.Background(), dbname, mock.NilOption)
	return &db{
		client: c.(*client),
		dbName: dbname,
	}, err
}
