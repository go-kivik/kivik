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

	"github.com/go-kivik/couchdb/v4/chttp"
	"github.com/go-kivik/kivik/v4/driver"
)

func TestVersion2(t *testing.T) {
	tests := []struct {
		name     string
		client   *client
		expected *driver.Version
		status   int
		err      string
	}{
		{
			name:   "network error",
			client: newTestClient(nil, errors.New("net error")),
			status: http.StatusBadGateway,
			err:    `Get "?http://example.com/"?: net error`,
		},
		{
			name: "invalid JSON response",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"couchdb":"Welcome","uuid":"a902efb0fac143c2b1f97160796a6347","version":"1.6.1","vendor":{"name":[]}}`)),
			}, nil),
			status: http.StatusBadGateway,
			err:    "json: cannot unmarshal array into Go ",
		},
		{
			name: "error response",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusInternalServerError,
				Body:       io.NopCloser(strings.NewReader("")),
			}, nil),
			status: http.StatusInternalServerError,
			err:    "Internal Server Error",
		},
		{
			name: "CouchDB 1.6.1",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"couchdb":"Welcome","uuid":"a902efb0fac143c2b1f97160796a6347","version":"1.6.1","vendor":{"version":"1.6.1","name":"The Apache Software Foundation"}}`)),
			}, nil),
			expected: &driver.Version{
				Version:     "1.6.1",
				Vendor:      "The Apache Software Foundation",
				RawResponse: []byte(`{"couchdb":"Welcome","uuid":"a902efb0fac143c2b1f97160796a6347","version":"1.6.1","vendor":{"version":"1.6.1","name":"The Apache Software Foundation"}}`),
			},
		},
		{
			name: "CouchDB 2.0.0",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"couchdb":"Welcome","version":"2.0.0","vendor":{"name":"The Apache Software Foundation"}}`)),
			}, nil),
			expected: &driver.Version{
				Version:     "2.0.0",
				Vendor:      "The Apache Software Foundation",
				RawResponse: []byte(`{"couchdb":"Welcome","version":"2.0.0","vendor":{"name":"The Apache Software Foundation"}}`),
			},
		},
		{
			name: "CouchDB 2.1.0",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"couchdb":"Welcome","version":"2.1.0","features":["scheduler"],"vendor":{"name":"The Apache Software Foundation"}}`)),
			}, nil),
			expected: &driver.Version{
				Version:     "2.1.0",
				Vendor:      "The Apache Software Foundation",
				Features:    []string{"scheduler"},
				RawResponse: []byte(`{"couchdb":"Welcome","version":"2.1.0","features":["scheduler"],"vendor":{"name":"The Apache Software Foundation"}}`),
			},
		},
		{
			name: "Cloudant 2017-10-23",
			client: newTestClient(&http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"couchdb":"Welcome","version":"2.0.0","vendor":{"name":"IBM Cloudant","version":"6365","variant":"paas"},"features":["geo","scheduler"]}`)),
			}, nil),
			expected: &driver.Version{
				Version:     "2.0.0",
				Vendor:      "IBM Cloudant",
				Features:    []string{"geo", "scheduler"},
				RawResponse: []byte(`{"couchdb":"Welcome","version":"2.0.0","vendor":{"name":"IBM Cloudant","version":"6365","variant":"paas"},"features":["geo","scheduler"]}`),
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.Version(context.Background())
			testy.StatusErrorRE(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestVersionConstant(t *testing.T) {
	if Version != chttp.Version {
		t.Errorf("CouchDB version (%s) and chttp version (%s) don't match",
			Version, chttp.Version)
	}
}
