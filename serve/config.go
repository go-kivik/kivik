package serve

import (
	"context"
	"net/http"

	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/serve/config/memconf"
)

// defaultConfig returns a default server configuration.
func defaultConfig() *config.Config {
	ctx := context.Background()
	conf := memconf.New()
	conf.SetContext(ctx, "log", "level", "info")
	conf.SetContext(ctx, "httpd", "enable_compression", "true")
	conf.SetContext(ctx, "httpd", "compression_level", "8")
	conf.SetContext(ctx, "httpd", "port", "5984")
	return config.New(conf)
}

func getConfig(w http.ResponseWriter, r *http.Request) error {
	conf, err := GetService(r).Config().GetAll()
	if err != nil {
		return err
	}
	return serveJSON(w, conf)
}

func getConfigSection(w http.ResponseWriter, r *http.Request) error {
	sec, ok := stringParam(r, "section")
	if !ok {
		return errors.Status(http.StatusBadRequest, "section required")
	}
	conf, err := GetService(r).Config().GetSection(sec)
	if err != nil {
		return err
	}
	return serveJSON(w, conf)
}

func getConfigItem(w http.ResponseWriter, r *http.Request) error {
	sec, ok := stringParam(r, "section")
	if !ok {
		return errors.Status(http.StatusBadRequest, "section required")
	}
	key, ok := stringParam(r, "key")
	if !ok {
		return errors.Status(http.StatusBadRequest, "key required")
	}
	conf, err := GetService(r).Config().Get(sec, key)
	if err != nil {
		return err
	}
	return serveJSON(w, conf)
}
