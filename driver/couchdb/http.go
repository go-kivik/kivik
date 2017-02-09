package couchdb

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const jsonType = "application/json"

func (c *client) getJSON(url string, i interface{}) error {
	fullURL := c.url(url)
	fmt.Printf("URL = %s\n", fullURL)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return err
	}
	req.Header.Add("Accept", jsonType)
	if c.authUser != "" {
		req.SetBasicAuth(c.authUser, c.authPass)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	if err := ResponseError(resp); err != nil {
		return err
	}
	if resp.Header.Get("Content-Type") != jsonType {
		return fmt.Errorf("Unexpected response type: %s", resp.Header.Get("Content-Type"))
	}
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	return dec.Decode(i)
}
