package proxydb

import (
	"bytes"
	"encoding/json"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

type rows struct {
	*kivik.ResultSet
}

var _ driver.Rows = &rows{}

func (r *rows) Next(row *driver.Row) error {
	if !r.ResultSet.Next() {
		return r.ResultSet.Err()
	}
	var value json.RawMessage
	if err := r.ResultSet.ScanValue(&value); err != nil {
		return err
	}
	var doc json.RawMessage
	if err := r.ResultSet.ScanDoc(&doc); err != nil {
		return err
	}
	var err error
	row.ID, err = r.ResultSet.ID()
	if err != nil {
		return err
	}
	key, err := r.ResultSet.Key()
	if err != nil {
		return err
	}
	row.Key = json.RawMessage(key)
	row.Value = bytes.NewReader(value)
	row.Doc = bytes.NewReader(doc)
	return nil
}

func (r *rows) Close() error {
	return r.ResultSet.Close()
}

func (r *rows) Offset() int64 {
	md, err := r.ResultSet.Metadata()
	if err != nil {
		return 0
	}
	return md.Offset
}

func (r *rows) TotalRows() int64 {
	md, err := r.ResultSet.Metadata()
	if err != nil {
		return 0
	}
	return md.TotalRows
}

func (r *rows) UpdateSeq() string {
	md, err := r.ResultSet.Metadata()
	if err != nil {
		return ""
	}
	return md.UpdateSeq
}
