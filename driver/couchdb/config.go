package couchdb

import "net/http"

// GetAllConfig returns the entire server config
func (c *client) GetAllConfig() (map[string]map[string]string, error) {
	var conf map[string]map[string]string
	return conf, c.newRequest(http.MethodGet, "_config").
		AddHeader("Accept", typeJSON).
		DoJSON(&conf)
}
