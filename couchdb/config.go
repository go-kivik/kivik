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
	"io"
	"net/http"
	"strings"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
)

// couch1ConfigNode can be passed to any of the Config-related methods as the
// node name, to query the /_config endpoint in a CouchDB 1.x-compatible way.
const couch1ConfigNode = ""

var _ driver.Configer = &client{}

func configURL(node string, parts ...string) string {
	var components []string
	if node == couch1ConfigNode {
		components = append(make([]string, 0, len(parts)+1),
			"_config")
	} else {
		components = append(make([]string, 0, len(parts)+3), // nolint:gomnd
			"_node", node, "_config",
		)
	}
	components = append(components, parts...)
	return "/" + strings.Join(components, "/")
}

func (c *client) Config(ctx context.Context, node string) (driver.Config, error) {
	cf := driver.Config{}
	err := c.Client.DoJSON(ctx, http.MethodGet, configURL(node), nil, &cf)
	return cf, err
}

func (c *client) ConfigSection(ctx context.Context, node, section string) (driver.ConfigSection, error) {
	sec := driver.ConfigSection{}
	err := c.Client.DoJSON(ctx, http.MethodGet, configURL(node, section), nil, &sec)
	return sec, err
}

func (c *client) ConfigValue(ctx context.Context, node, section, key string) (string, error) {
	var value string
	err := c.Client.DoJSON(ctx, http.MethodGet, configURL(node, section, key), nil, &value)
	return value, err
}

func (c *client) SetConfigValue(ctx context.Context, node, section, key, value string) (string, error) {
	body, _ := json.Marshal(value) // Strings never cause JSON marshaling errors
	var old string
	opts := &chttp.Options{
		Body: io.NopCloser(bytes.NewReader(body)),
	}
	err := c.Client.DoJSON(ctx, http.MethodPut, configURL(node, section, key), opts, &old)
	return old, err
}

func (c *client) DeleteConfigKey(ctx context.Context, node, section, key string) (string, error) {
	var value string
	err := c.Client.DoJSON(ctx, http.MethodDelete, configURL(node, section, key), nil, &value)
	return value, err
}
