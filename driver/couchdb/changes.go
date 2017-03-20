package couchdb

import (
	"encoding/json"
	"io"

	"github.com/flimzy/kivik/driver"
)

type changesRows struct {
	body   io.ReadCloser
	dec    *json.Decoder
	closed bool
}

func newChangesRows(r io.ReadCloser) *changesRows {
	return &changesRows{
		body: r,
		dec:  json.NewDecoder(r),
	}
}

var _ driver.Rows = &changesRows{}

func (r *changesRows) Close() error {
	return r.body.Close()
}

func (r *changesRows) Next(row *driver.Row) error {
	if r.closed {
		return io.EOF
	}
	if !r.dec.More() {
		return io.EOF
	}

	return r.dec.Decode(row)
}

func (r *changesRows) Offset() int64     { return 0 }
func (r *changesRows) TotalRows() int64  { return 0 }
func (r *changesRows) UpdateSeq() string { return "" }
