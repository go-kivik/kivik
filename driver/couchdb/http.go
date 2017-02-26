package couchdb

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
)

const jsonType = "application/json"
const textType = "text/plain"

func (c *client) makeRequest(method string, url string, query url.Values, accept string) (*http.Response, error) {
	fullURL := c.url(url, query)
	req, err := http.NewRequest(method, fullURL, nil)
	if err != nil {
		return nil, err
	}
	var reqMediaType string
	if accept != "" {
		reqMediaType, _, err = mime.ParseMediaType(accept)
		if err != nil {
			return nil, fmt.Errorf("Invalid Accept type: %s", accept)
		}
		req.Header.Add("Accept", accept)
	}
	if c.auth != nil {
		if err = c.auth.Authenticate(req); err != nil {
			return nil, err
		}
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	if err = ResponseError(resp); err != nil {
		return nil, err
	}
	if accept != "" {
		respMediaType, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type"))
		if err != nil {
			return nil, fmt.Errorf("failed to parse header Content-Type: %s", resp.Header.Get("Content-Type"))
		}
		if respMediaType != reqMediaType {
			return nil, fmt.Errorf("Unexpected response type '%s'", respMediaType)
		}
	}
	return resp, nil
}

func (c *client) doJSON(method, url string, i interface{}, query url.Values) error {
	resp, err := c.makeRequest(method, url, query, jsonType)
	if err != nil {
		return err
	}
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	return dec.Decode(i)
}

func (c *client) getJSON(url string, i interface{}, query url.Values) error {
	return c.doJSON(http.MethodGet, url, i, query)
}

func (c *client) putJSON(url string, i interface{}, query url.Values) error {
	return c.doJSON(http.MethodPut, url, i, query)
}

func (c *client) deleteJSON(url string, i interface{}, query url.Values) error {
	return c.doJSON(http.MethodDelete, url, i, query)
}

func (c *client) getText(url string, buf []byte, query url.Values) error {
	resp, err := c.makeRequest("GET", url, query, textType)
	if err != nil {
		return err
	}
	_, err = resp.Body.Read(buf)
	if err == nil || err == io.EOF {
		return nil
	}
	return err
}

func (c *client) head(url string, query url.Values) error {
	_, err := c.makeRequest("HEAD", url, query, "")
	return err
}
