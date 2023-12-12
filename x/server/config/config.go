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

// Package config manages server configuration.
package config

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4/internal"
)

// Config provides access to read server configuration. Configuration
// backends that allow modifying configuraiont will also implement
// [ConfigWriter].
type Config interface {
	All(context.Context) (map[string]interface{}, error)
	Section(context.Context, string) (map[string]interface{}, error)
	Key(context.Context, string, string) (interface{}, error)
	Reload(context.Context) error
}

// Writer allows setting server configuration.
type Writer interface {
	SetSection(context.Context, string, interface{}) error
	SetKey(context.Context, string, string, interface{}) error
	Delete(context.Context, string, string) error
}

type defaultConfig struct {
	conf map[string]interface{}
}

var _ Config = (*defaultConfig)(nil)

// Default returns a Config implementation which returns read-only default
// values for all configuration options.
func Default() Config {
	return &defaultConfig{
		conf: map[string]interface{}{
			"couchdb": map[string]interface{}{
				"users_db_suffix": "_users",
			},
		},
	}
}

func (c *defaultConfig) All(context.Context) (map[string]interface{}, error) {
	return c.conf, nil
}

func (c *defaultConfig) Section(_ context.Context, section string) (map[string]interface{}, error) {
	s, _ := c.conf[section].(map[string]interface{})
	return s, nil
}

func (c *defaultConfig) Key(_ context.Context, section, key string) (interface{}, error) {
	s, _ := c.conf[section].(map[string]interface{})

	if v, ok := s[key]; ok {
		return v, nil
	}
	return nil, &internal.Error{Status: http.StatusNotFound, Message: "unknown_config_value"}
}

func (c *defaultConfig) Reload(context.Context) error {
	return nil
}
