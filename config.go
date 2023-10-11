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

	"github.com/go-kivik/kivik/v4/driver"
)

// Config represents all the config sections.
type Config map[string]ConfigSection

// ConfigSection represents all key/value pairs for a section of configuration.
type ConfigSection map[string]string

// Config returns the entire [server config], for the specified node.
//
// [server config]: http://docs.couchdb.org/en/stable/api/server/configuration.html#get--_node-node-name-_config
func (c *Client) Config(ctx context.Context, node string) (Config, error) {
	endQuery, err := c.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
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
	return nil, errConfigNotImplemented
}

// ConfigSection returns the requested server [config section] for the specified node.
//
// [section]: http://docs.couchdb.org/en/stable/api/server/configuration.html#node-node-name-config-section
func (c *Client) ConfigSection(ctx context.Context, node, section string) (ConfigSection, error) {
	endQuery, err := c.startQuery()
	if err != nil {
		return nil, err
	}
	defer endQuery()
	if configer, ok := c.driverClient.(driver.Configer); ok {
		sec, err := configer.ConfigSection(ctx, node, section)
		return ConfigSection(sec), err
	}
	return nil, errConfigNotImplemented
}

// ConfigValue returns a single [config value] for the specified node.
//
// [config value]: http://docs.couchdb.org/en/stable/api/server/configuration.html#get--_node-node-name-_config-section-key
func (c *Client) ConfigValue(ctx context.Context, node, section, key string) (string, error) {
	endQuery, err := c.startQuery()
	if err != nil {
		return "", err
	}
	defer endQuery()
	if configer, ok := c.driverClient.(driver.Configer); ok {
		return configer.ConfigValue(ctx, node, section, key)
	}
	return "", errConfigNotImplemented
}

// SetConfigValue sets the server's [config value] on the specified node, creating
// the key if it doesn't exist. It returns the old value.
//
// [config value]: http://docs.couchdb.org/en/stable/api/server/configuration.html#put--_node-node-name-_config-section-key
func (c *Client) SetConfigValue(ctx context.Context, node, section, key, value string) (string, error) {
	endQuery, err := c.startQuery()
	if err != nil {
		return "", err
	}
	defer endQuery()
	if configer, ok := c.driverClient.(driver.Configer); ok {
		return configer.SetConfigValue(ctx, node, section, key, value)
	}
	return "", errConfigNotImplemented
}

// DeleteConfigKey deletes the [configuration key] and associated value from the
// specified node. It returns the old value.
//
// [configuration key]: http://docs.couchdb.org/en/stable/api/server/configuration.html#delete--_node-node-name-_config-section-key
func (c *Client) DeleteConfigKey(ctx context.Context, node, section, key string) (string, error) {
	endQuery, err := c.startQuery()
	if err != nil {
		return "", err
	}
	defer endQuery()
	if configer, ok := c.driverClient.(driver.Configer); ok {
		return configer.DeleteConfigKey(ctx, node, section, key)
	}
	return "", errConfigNotImplemented
}
