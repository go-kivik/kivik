package memory

import (
	"context"
	"errors"

	"github.com/flimzy/kivik/driver"
)

var errFindNotImplemented = errors.New("find feature not yet implemented")

func (d *db) Find(_ context.Context, query interface{}) (driver.Rows, error) {
	return nil, nil
}

func (d *db) CreateIndex(_ context.Context, ddoc, name string, index interface{}) error {
	return errFindNotImplemented
}

func (d *db) GetIndexes(_ context.Context) ([]driver.Index, error) {
	return nil, errFindNotImplemented
}

func (d *db) DeleteIndex(_ context.Context, ddoc, name string) error {
	return errFindNotImplemented
}
