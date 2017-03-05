package serve

import (
	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/serve/config/memconf"
)

// defaultConfig returns a default server configuration.
func defaultConfig() *config.Config {
	conf := memconf.New()
	conf.Set("httpd", "enable_compression", "true")
	conf.Set("httpd", "compression_level", "8")
	return config.New(conf)
}
