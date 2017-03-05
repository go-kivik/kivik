package couchdb

import (
	"fmt"
	"net/http"

	"github.com/flimzy/kivik/driver"
)

func (c *client) Config() (driver.Config, error) {
	return &config{client: c}, nil
}

type config struct {
	*client
}

var _ driver.Config = &config{}

func (c *config) GetAll() (map[string]map[string]string, error) {
	conf := map[string]map[string]string{}
	return conf, c.newRequest(http.MethodGet, "/_config").
		AddHeader("Accept", typeJSON).
		DoJSON(&conf)
}

func (c *config) Set(secName, key, value string) error {
	_, err := c.newRequest(http.MethodPut, fmt.Sprintf("/_config/%s/%s", secName, key)).
		AddHeader("Content-Type", typeJSON).
		AddHeader("Accept", typeJSON).
		BodyJSON(value).
		Do()
	return err
}

func (c *config) Delete(secName, key string) error {
	_, err := c.newRequest(http.MethodDelete, fmt.Sprintf("/_config/%s/%s", secName, key)).
		Do()
	return err
}
