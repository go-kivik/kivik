package couchdb

import (
	"encoding/json"
	"fmt"
	"mime"
	"net/http"
	"net/url"
)

const jsonType = "application/json"
const textType = "text/plain"

func (c *client) makeRequest(url string, query url.Values, accept string) (*http.Response, error) {
	fullURL := c.url(url, query)
	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, err
	}
	reqMediaType, _, err := mime.ParseMediaType(accept)
	if err != nil {
		return nil, fmt.Errorf("Invalid Accept type: %s", accept)
	}
	req.Header.Add("Accept", accept)
	if c.authUser != "" {
		req.SetBasicAuth(c.authUser, c.authPass)
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if err = ResponseError(resp); err != nil {
		return nil, err
	}
	respMediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse header Content-Type: %s", resp.Header.Get("Content-Type"))
	}
	if respMediaType != reqMediaType {
		return nil, fmt.Errorf("Unexpected response type '%s'", respMediaType)
	}
	return resp, nil
}

func (c *client) getJSON(url string, i interface{}, query url.Values) error {
	resp, err := c.makeRequest(url, query, jsonType)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	return dec.Decode(i)
}

func (c *client) getText(url string, buf []byte, query url.Values) error {
	resp, err := c.makeRequest(url, query, textType)
	if err != nil {
		return err
	}
	_, err = resp.Body.Read(buf)
	return err
}
