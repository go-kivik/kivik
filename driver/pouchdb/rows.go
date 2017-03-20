package pouchdb

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/kivik/driver"
)

var jsJSON = js.Global.Get("JSON")

type rows struct {
	*js.Object
	Off   int64 `js:"offset"`
	TRows int64 `js:"total_rows"`
}

var _ driver.Rows = &rows{}

func (r *rows) Close() error {
	r.Set("rows", nil) // Free up memory used by any remaining rows
	return nil
}

func (r *rows) Next(row *driver.Row) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	if r.Get("rows") == js.Undefined || r.Get("rows").Length() == 0 {
		return io.EOF
	}
	next := r.Get("rows").Call("shift")
	row.ID = next.Get("id").String()
	row.Key = next.Get("key").String()
	row.Value = json.RawMessage(jsJSON.Call("stringify", next.Get("value")).String())
	if doc := next.Get("doc"); doc != js.Undefined {
		row.Doc = json.RawMessage(jsJSON.Call("stringify", doc).String())
	}
	return nil
}

func (r *rows) Offset() int64 {
	return r.Off
}

func (r *rows) TotalRows() int64 {
	return r.TRows
}
