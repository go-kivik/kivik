package couchdb

import (
	"errors"
	"fmt"
	"io"

	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

// Log returns server logs.
func (c *client) Log(length, offset int64) (io.ReadCloser, error) {
	if length < 0 {
		return nil, errors.New("invalid length specified")
	}
	if offset < 0 {
		return nil, errors.New("invalid offset specified")
	}
	resp, err := c.DoReq(chttp.MethodGet, fmt.Sprintf("/_log?bytes=%d&offset=%d", length, offset), nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, chttp.ResponseError(resp.Response)
}
