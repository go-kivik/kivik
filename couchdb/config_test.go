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
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/pkg/errors"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
)

func TestConfig(t *testing.T) {
	type tst struct {
		client   *client
		node     string
		expected driver.Config
		status   int
		err      string
	}
	tests := testy.NewTable()
	tests.Add("network error", tst{
		client: newTestClient(nil, errors.New("net error")),
		node:   "local",
		status: http.StatusBadGateway,
		err:    `^Get "?http://example.com/_node/local/_config"?: net error$`,
	})
	tests.Add("Couch 1.x path", tst{
		client: newTestClient(nil, errors.New("net error")),
		node:   Couch1ConfigNode,
		status: http.StatusBadGateway,
		err:    `^Get "?http://example.com/_config"?: net error$`,
	})
	tests.Add("success", tst{
		client: newTestClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"foo":{"asd":"baz"}}`)),
		}, nil),
		node: "local",
		expected: driver.Config{
			"foo": driver.ConfigSection{"asd": "baz"},
		},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.Config(context.Background(), test.node)
		testy.StatusErrorRE(t, test.err, test.status, err)
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestConfigSection(t *testing.T) {
	type tst struct {
		client        *client
		node, section string
		expected      driver.ConfigSection
		status        int
		err           string
	}
	tests := testy.NewTable()
	tests.Add("network error", tst{
		client:  newTestClient(nil, errors.New("net error")),
		node:    "local",
		section: "foo",
		status:  http.StatusBadGateway,
		err:     `^Get "?http://example.com/_node/local/_config/foo"?: net error$`,
	})
	tests.Add("Couch 1.x path", tst{
		client:  newTestClient(nil, errors.New("net error")),
		node:    Couch1ConfigNode,
		section: "foo",
		status:  http.StatusBadGateway,
		err:     `^Get "?http://example.com/_config/foo"?: net error$`,
	})
	tests.Add("success", tst{
		client: newTestClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`{"fds":"baz"}`)),
		}, nil),
		node:     "local",
		section:  "foo",
		expected: driver.ConfigSection{"fds": "baz"},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.ConfigSection(context.Background(), test.node, test.section)
		testy.StatusErrorRE(t, test.err, test.status, err)
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestConfigValue(t *testing.T) {
	type tst struct {
		client             *client
		node, section, key string
		expected           string
		status             int
		err                string
	}
	tests := testy.NewTable()
	tests.Add("network error", tst{
		client:  newTestClient(nil, errors.New("net error")),
		node:    "local",
		section: "foo",
		key:     "tre",
		status:  http.StatusBadGateway,
		err:     `Get "?http://example.com/_node/local/_config/foo/tre"?: net error`,
	})
	tests.Add("Couch 1.x path", tst{
		client:  newTestClient(nil, errors.New("net error")),
		node:    Couch1ConfigNode,
		section: "foo",
		key:     "bar",
		status:  http.StatusBadGateway,
		err:     `Get "?http://example.com/_config/foo/bar"?: net error`,
	})
	tests.Add("success", tst{
		client: newTestClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`"baz"`)),
		}, nil),
		node:     "local",
		section:  "foo",
		key:      "bar",
		expected: "baz",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.ConfigValue(context.Background(), test.node, test.section, test.key)
		testy.StatusErrorRE(t, test.err, test.status, err)
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestSetConfigValue(t *testing.T) {
	type tst struct {
		client                    *client
		node, section, key, value string
		expected                  string
		status                    int
		err                       string
	}
	tests := testy.NewTable()
	tests.Add("network error", tst{
		client:  newTestClient(nil, errors.New("net error")),
		node:    "local",
		section: "foo",
		key:     "bar",
		status:  http.StatusBadGateway,
		err:     `Put "?http://example.com/_node/local/_config/foo/bar"?: net error`,
	})
	tests.Add("Couch 1.x path", tst{
		client:  newTestClient(nil, errors.New("net error")),
		node:    Couch1ConfigNode,
		section: "foo",
		key:     "bar",
		status:  http.StatusBadGateway,
		err:     `^Put "?http://example.com/_config/foo/bar"?: net error$`,
	})
	tests.Add("success", tst{
		client: newCustomClient(func(r *http.Request) (*http.Response, error) {
			var val string
			defer r.Body.Close() // nolint: errcheck
			if err := json.NewDecoder(r.Body).Decode(&val); err != nil {
				return nil, err
			}
			if val != "baz" {
				return nil, errors.Errorf("Unexpected value: %s", val)
			}

			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`"old"`)),
			}, nil
		}),
		node:     "local",
		section:  "foo",
		key:      "bar",
		value:    "baz",
		expected: "old",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.SetConfigValue(context.Background(), test.node, test.section, test.key, test.value)
		testy.StatusErrorRE(t, test.err, test.status, err)
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestDeleteConfigKey(t *testing.T) {
	type tst struct {
		client             *client
		node, section, key string
		expected           string
		status             int
		err                string
	}
	tests := testy.NewTable()
	tests.Add("network error", tst{
		client:  newTestClient(nil, errors.New("net error")),
		node:    "local",
		section: "foo",
		key:     "bar",
		status:  http.StatusBadGateway,
		err:     `Delete "?http://example.com/_node/local/_config/foo/bar"?: net error`,
	})
	tests.Add("Couch 1.x path", tst{
		client:  newTestClient(nil, errors.New("net error")),
		node:    Couch1ConfigNode,
		section: "foo",
		key:     "bar",
		status:  http.StatusBadGateway,
		err:     `Delete "?http://example.com/_config/foo/bar"?: net error`,
	})
	tests.Add("success", tst{
		client: newTestClient(&http.Response{
			StatusCode: http.StatusOK,
			Body:       io.NopCloser(strings.NewReader(`"old"`)),
		}, nil),
		node:     "local",
		section:  "foo",
		key:      "bar",
		expected: "old",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.DeleteConfigKey(context.Background(), test.node, test.section, test.key)
		testy.StatusErrorRE(t, test.err, test.status, err)
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}
