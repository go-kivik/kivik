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
	All(context.Context) (map[string]map[string]string, error)
	Section(context.Context, string) (map[string]string, error)
	Key(context.Context, string, string) (string, error)
	Reload(context.Context) error
}

// Writer allows setting server configuration.
type Writer interface {
	// SetKey sets a new config value, and returns the old value.
	SetKey(context.Context, string, string, string) (string, error)
	Delete(context.Context, string, string) (string, error)
}

type defaultConfig struct {
	conf map[string]map[string]string
}

var (
	_ Config = (*defaultConfig)(nil)
	_ Writer = (*defaultConfig)(nil)
)

// Default returns a Config implementation which returns default values for all
// configuration options, and preserves changed settings only until restart.
func Default() Config {
	return &defaultConfig{
		conf: map[string]map[string]string{
			"couchdb": {
				"users_db_suffix": "_users",
			},
		},
	}
}

// Map returns a Config implementation which returns the provided configuration.
func Map(conf map[string]map[string]string) Config {
	return &defaultConfig{conf: conf}
}

func (c *defaultConfig) All(context.Context) (map[string]map[string]string, error) {
	return c.conf, nil
}

func (c *defaultConfig) Section(_ context.Context, section string) (map[string]string, error) {
	return c.conf[section], nil
}

func (c *defaultConfig) Key(_ context.Context, section, key string) (string, error) {
	if v, ok := c.conf[section][key]; ok {
		return v, nil
	}
	return "", &internal.Error{Status: http.StatusNotFound, Message: "unknown_config_value"}
}

func (c *defaultConfig) Reload(context.Context) error {
	return nil
}

func (c *defaultConfig) SetKey(_ context.Context, section, key string, value string) (string, error) {
	s, ok := c.conf[section]
	if !ok {
		s = map[string]string{}
		c.conf[section] = s
	}
	old := s[key]
	s[key] = value
	return old, nil
}

func (c *defaultConfig) Delete(_ context.Context, section, key string) (string, error) {
	if v, ok := c.conf[section][key]; ok {
		delete(c.conf[section], key)
		return v, nil
	}
	return "", &internal.Error{Status: http.StatusNotFound, Message: "unknown_config_value"}
}
