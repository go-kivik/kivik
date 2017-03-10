package couchdb

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/flimzy/kivik/driver/couchdb/chttp"
)

// LogContext returns server logs.
func (c *client) LogContext(ctx context.Context, length, offset int64) (io.ReadCloser, error) {
	if length < 0 {
		return nil, errors.New("invalid length specified")
	}
	if offset < 0 {
		return nil, errors.New("invalid offset specified")
	}
	resp, err := c.DoReq(ctx, chttp.MethodGet, fmt.Sprintf("/_log?bytes=%d&offset=%d", length, offset), nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, chttp.ResponseError(resp.Response)
}
