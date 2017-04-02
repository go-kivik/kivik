package couchdb

import (
	"context"
	"fmt"
	"io"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
	"github.com/flimzy/kivik/errors"
)

// LogContext returns server logs.
func (c *client) LogContext(ctx context.Context, length, offset int64) (io.ReadCloser, error) {
	if length < 0 {
		return nil, errors.Status(kivik.StatusBadRequest, "invalid length specified")
	}
	if offset < 0 {
		return nil, errors.Status(kivik.StatusBadRequest, "invalid offset specified")
	}
	resp, err := c.DoReq(ctx, kivik.MethodGet, fmt.Sprintf("/_log?bytes=%d&offset=%d", length, offset), nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, chttp.ResponseError(resp)
}
