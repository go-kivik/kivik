package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
)

func (c *client) HTTPRequest(method, path string, body io.Reader) (*http.Request, *http.Client, error) {
	parsedURL, err := url.Parse(path)
	if err != nil {
		return nil, nil, err
	}
	fullURL := c.url(parsedURL.Path, parsedURL.Query())
	req, err := http.NewRequest(method, fullURL, body)
	return req, c.httpClient, err
}

const (
	typeJSON  = "application/json"
	typeText  = "text/plain"
	typeMixed = "multipart/mixed"
)

type request struct {
	client *client
	method string
	url    string
	body   io.Reader
	query  url.Values
	header http.Header
}

func (c *client) newRequest(method, url string) *request {
	return &request{
		client: c,
		method: method,
		url:    url,
		header: http.Header{},
	}
}

func (r *request) BodyJSON(i interface{}) *request {
	body, err := json.Marshal(i)
	if err != nil {
		panic(err)
	}
	r.body = bytes.NewReader(body)
	return r
}

func (r *request) Body(body io.Reader) *request {
	r.body = body
	return r
}

func (r *request) Query(query url.Values) *request {
	r.query = query
	return r
}

func (r *request) AddHeader(key, value string) *request {
	r.header.Add(key, value)
	return r
}

func (r *request) If(cond bool, fn func(*request)) *request {
	if cond {
		fn(r)
	}
	return r
}

func (r *request) Do() (*http.Response, error) {
	c := r.client
	fullURL := c.url(r.url, r.query)
	req, err := http.NewRequest(r.method, fullURL, r.body)
	if err != nil {
		return nil, err
	}
	var reqMediaType string
	if accept := r.header.Get("Accept"); accept != "" {
		reqMediaType, _, err = mime.ParseMediaType(accept)
		if err != nil {
			return nil, fmt.Errorf("Invalid Accept type: %s", accept)
		}
	}
	// Copy all the headers
	for key, values := range r.header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if err = ResponseError(resp); err != nil {
		return nil, err
	}
	if reqMediaType != "" {
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

func (r *request) DoJSON(i interface{}) error {
	resp, err := r.Do()
	if err != nil {
		return err
	}
	dec := json.NewDecoder(resp.Body)
	defer resp.Body.Close()
	return dec.Decode(i)
}

func (c *client) makeRequest(method string, url string, query url.Values, accept string) (*http.Response, error) {
	return c.newRequest(method, url).
		AddHeader("Accept", accept).
		Query(query).
		Do()
}

func (c *client) doJSON(method, url string, i interface{}, query url.Values) error {
	resp, err := c.makeRequest(method, url, query, typeJSON)
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
	resp, err := c.makeRequest("GET", url, query, typeText)
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
