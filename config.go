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

package kivik

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
)

// Config represents all the config sections.
//
// Note that the Config struct, and all of the config-related methods are
// considered experimental, and may change in the future.
type Config map[string]ConfigSection

// ConfigSection represents all key/value pairs for a section of configuration.
type ConfigSection map[string]string

var configNotImplemented = &Error{HTTPStatus: http.StatusNotImplemented, Message: "kivik: driver does not support Config interface"}

// Config returns the entire server config, for the specified node.
//
// See http://docs.couchdb.org/en/stable/api/server/configuration.html#get--_node-node-name-_config
func (c *Client) Config(ctx context.Context, node string) (Config, error) {
	if configer, ok := c.driverClient.(driver.Configer); ok {
		driverCf, err := configer.Config(ctx, node)
		if err != nil {
			return nil, err
		}
		cf := Config{}
		for k, v := range driverCf {
			cf[k] = ConfigSection(v)
		}
		return cf, nil
	}
	return nil, configNotImplemented
}

// ConfigSection returns the requested section of the server config for the
// specified node.
//
// See http://docs.couchdb.org/en/stable/api/server/configuration.html#node-node-name-config-section
func (c *Client) ConfigSection(ctx context.Context, node, section string) (ConfigSection, error) {
	if configer, ok := c.driverClient.(driver.Configer); ok {
		sec, err := configer.ConfigSection(ctx, node, section)
		return ConfigSection(sec), err
	}
	return nil, configNotImplemented
}

// ConfigValue returns a single config value for the specified node.
//
// See http://docs.couchdb.org/en/stable/api/server/configuration.html#get--_node-node-name-_config-section-key
func (c *Client) ConfigValue(ctx context.Context, node, section, key string) (string, error) {
	if configer, ok := c.driverClient.(driver.Configer); ok {
		return configer.ConfigValue(ctx, node, section, key)
	}
	return "", configNotImplemented
}

// SetConfigValue sets the server's config value on the specified node, creating
// the key if it doesn't exist. It returns the old value.
//
// See http://docs.couchdb.org/en/stable/api/server/configuration.html#put--_node-node-name-_config-section-key
func (c *Client) SetConfigValue(ctx context.Context, node, section, key, value string) (string, error) {
	if configer, ok := c.driverClient.(driver.Configer); ok {
		return configer.SetConfigValue(ctx, node, section, key, value)
	}
	return "", configNotImplemented
}

// DeleteConfigKey deletes the configuration key and associated value from the
// specified node. It returns the old value.
//
// See http://docs.couchdb.org/en/stable/api/server/configuration.html#delete--_node-node-name-_config-section-key
func (c *Client) DeleteConfigKey(ctx context.Context, node, section, key string) (string, error) {
	if configer, ok := c.driverClient.(driver.Configer); ok {
		return configer.DeleteConfigKey(ctx, node, section, key)
	}
	return "", configNotImplemented
}
