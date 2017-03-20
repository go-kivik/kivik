package couchdb

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"github.com/flimzy/kivik/driver"
)

type rows struct {
	offset    int64
	totalRows int64
	updateSeq string
	body      io.ReadCloser
	dec       *json.Decoder
	// closed is true after all rows have been processed
	closed bool
}

var _ driver.Rows = &rows{}

func newRows(r io.ReadCloser) *rows {
	return &rows{
		body: r,
	}
}

func (r *rows) Offset() int64 {
	return r.offset
}

func (r *rows) TotalRows() int64 {
	return r.totalRows
}

func (r *rows) UpdateSeq() string {
	return r.updateSeq
}

func (r *rows) Close() error {
	return r.body.Close()
}

func (r *rows) Next(row *driver.Row) error {
	if r.closed {
		return io.EOF
	}
	if r.dec == nil {
		// We haven't begun yet
		r.dec = json.NewDecoder(r.body)
		// consume the first '{'
		if err := r.consumeDelim(json.Delim('{')); err != nil {
			return err
		}
		if err := r.begin(); err != nil {
			return err
		}
	}

	err := r.nextRow(row)
	if err != nil {
		r.closed = true
		if err == io.EOF {
			return r.finish()
		}
	}
	return err
}

// begin parses the top-level of the result object; until rows
func (r *rows) begin() error {
	for {
		t, err := r.dec.Token()
		if err != nil {
			// I can't find a test case to trigger this, so it remains uncovered.
			return err
		}
		key, ok := t.(string)
		if !ok {
			// The JSON parser should never permit this
			return fmt.Errorf("Unexpected token: (%T) %v", t, t)
		}
		if key == "rows" {
			// Consume the first '['
			return r.consumeDelim(json.Delim('['))
		}
		if err := r.parseMeta(key); err != nil {
			return err
		}
	}
}

func (r *rows) finish() error {
	for {
		t, err := r.dec.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case json.Delim:
			if v != json.Delim('}') {
				// This should never happen, as the JSON parser should prevent it.
				return fmt.Errorf("Unexpected JSON delimiter: %c", v)
			}
		case string:
			if err := r.parseMeta(v); err != nil {
				return err
			}
		default:
			// This should never happen, as the JSON parser would never get
			// this far.
			return fmt.Errorf("Unexpected JSON token: (%T) '%s'", t, t)
		}
	}
}

// parseMeta parses result metadata
func (r *rows) parseMeta(key string) error {
	switch key {
	case "update_seq":
		return r.readUpdateSeq()
	case "offset":
		return r.dec.Decode(&r.offset)
	case "total_rows":
		return r.dec.Decode(&r.totalRows)
	}
	return fmt.Errorf("Unexpected key: %s", key)
}

func (r *rows) readUpdateSeq() error {
	var raw json.RawMessage
	if err := r.dec.Decode(&raw); err != nil {
		return err
	}
	r.updateSeq = string(bytes.Trim(raw, `""`))
	return nil
}

func (r *rows) nextRow(row *driver.Row) error {
	if !r.dec.More() {
		if err := r.consumeDelim(json.Delim(']')); err != nil {
			return err
		}
		return io.EOF
	}
	err := r.dec.Decode(row)
	return err
}

// consumeDelim consumes the expected delimiter from the stream, or returns an
// error if an unexpected token was found.
func (r *rows) consumeDelim(expectedDelim json.Delim) error {
	t, err := r.dec.Token()
	if err != nil {
		return err
	}
	d, ok := t.(json.Delim)
	if !ok {
		return fmt.Errorf("Unexpected token %T: %v", t, t)
	}
	if d != expectedDelim {
		return fmt.Errorf("Unexpected JSON delimiter: %c", d)
	}
	return nil
}
