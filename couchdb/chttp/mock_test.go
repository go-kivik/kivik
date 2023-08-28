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

package chttp

import (
	"io"
	"net/http"
	"net/url"
	"strings"
)

type customTransport func(*http.Request) (*http.Response, error)

var _ http.RoundTripper = customTransport(nil)

func (c customTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return c(req)
}

func newCustomClient(dsn string, fn func(*http.Request) (*http.Response, error)) *Client {
	c := &Client{
		Client: &http.Client{
			Transport: customTransport(fn),
		},
	}
	var err error
	if dsn == "" {
		dsn = "http://example.com/"
	}
	c.dsn, err = url.Parse(dsn)
	if err != nil {
		panic(err)
	}
	c.basePath = strings.TrimSuffix(c.dsn.Path, "/")
	return c
}

func newTestClient(resp *http.Response, err error) *Client {
	return newCustomClient("", func(_ *http.Request) (*http.Response, error) {
		return resp, err
	})
}

func Body(str string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(str))
}
