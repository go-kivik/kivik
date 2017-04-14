package pouchdb

import (
	"context"
	"encoding/json"
	"io"

	"github.com/gopherjs/gopherjs/js"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
)

var _ driver.Finder = &db{}

// buildIndex merges the ddoc and name into the index structure, as reqiured
// by the PouchDB-find plugin.
func buildIndex(ddoc, name string, index interface{}) (*js.Object, error) {
	i, err := bindings.Objectify(index)
	if err != nil {
		return nil, err
	}
	o := js.Global.Get("Object").New(i)
	if ddoc != "" {
		o.Set("ddoc", ddoc)
	}
	if name != "" {
		o.Set("name", name)
	}
	return o, nil
}

func (d *db) CreateIndexContext(ctx context.Context, ddoc, name string, index interface{}) error {
	indexObj, err := buildIndex(ddoc, name, index)
	if err != nil {
		return err
	}
	_, err = d.db.CreateIndex(ctx, indexObj)
	return err
}

func (d *db) FindContext(ctx context.Context, query interface{}) (driver.Rows, error) {
	result, err := d.db.Find(ctx, query)
	if err != nil {
		return nil, err
	}
	return &findRows{
		Object: result,
	}, nil
}

type findRows struct {
	*js.Object
}

var _ driver.Rows = &findRows{}

func (r *findRows) Offset() int64     { return 0 }
func (r *findRows) TotalRows() int64  { return 0 }
func (r *findRows) UpdateSeq() string { return "" }

func (r *findRows) Close() error {
	r.Delete("docs") // Free up memory used by any remaining rows
	return nil
}

func (r *findRows) Next(row *driver.Row) (err error) {
	defer bindings.RecoverError(&err)
	if r.Get("docs") == js.Undefined || r.Get("docs").Length() == 0 {
		return io.EOF
	}
	next := r.Get("docs").Call("shift")
	row.Doc = json.RawMessage(jsJSON.Call("stringify", next).String())
	return nil
}
