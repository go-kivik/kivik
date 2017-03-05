// Package config provides default server configuration.
package config

import (
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/serve/config/memconf"
)

// DefaultConfig returns a default server configuration.
func DefaultConfig() driver.Config {
	conf := memconf.New()
	conf.Set("httpd", "compression_enabled", "true")
	conf.Set("httpd", "compression_level", "8")
	return conf
}
