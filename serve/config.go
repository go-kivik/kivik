package serve

import (
	"encoding/json"
	"net/http"

	"github.com/flimzy/kivik/config"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/serve/config/memconf"
)

// defaultConfig returns a default server configuration.
func defaultConfig() *config.Config {
	conf := memconf.New()
	conf.Set("httpd", "enable_compression", "true")
	conf.Set("httpd", "compression_level", "8")
	return config.New(conf)
}

func getConfig(w http.ResponseWriter, r *http.Request) error {
	config, err := getClient(r).Config()
	if err != nil {
		return err
	}
	conf, err := config.GetAll()
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(conf)
}

func getConfigSection(w http.ResponseWriter, r *http.Request) error {
	config, err := getClient(r).Config()
	if err != nil {
		return err
	}
	sec, ok := stringParam(r, "section")
	if !ok {
		return errors.Status(http.StatusBadRequest, "section required")
	}
	conf, err := config.GetSection(sec)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(conf)
}

func getConfigItem(w http.ResponseWriter, r *http.Request) error {
	config, err := getClient(r).Config()
	if err != nil {
		return err
	}
	sec, ok := stringParam(r, "section")
	if !ok {
		return errors.Status(http.StatusBadRequest, "section required")
	}
	key, ok := stringParam(r, "key")
	if !ok {
		return errors.Status(http.StatusBadRequest, "key required")
	}
	conf, err := config.Get(sec, key)
	if err != nil {
		return err
	}
	return json.NewEncoder(w).Encode(conf)
}
