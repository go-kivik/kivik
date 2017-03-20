package proxy

import (
	"encoding/json"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
)

type rows struct {
	*kivik.Rows
}

var _ driver.Rows = &rows{}

func (r *rows) Next(row *driver.Row) error {
	if !r.Rows.Next() {
		return r.Rows.Err()
	}
	var value json.RawMessage
	if err := r.Rows.Scan(&value); err != nil {
		return err
	}
	row.ID = r.Rows.ID()
	row.Key = r.Rows.Key()
	row.Value = value
	return nil
}
