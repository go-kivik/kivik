package fs

import (
	"context"

	"github.com/flimzy/kivik/driver"
)

type db struct {
	*client
	dbName string
}

func (d *db) AllDocsContext(_ context.Context, docs interface{}, _ map[string]interface{}) (offset, totalrows int, seq string, err error) {
	// FIXME: Unimplemented
	return 0, 0, "", nil
}

func (d *db) GetContext(_ context.Context, docID string, doc interface{}, opts map[string]interface{}) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) CreateDocContext(_ context.Context, doc interface{}) (docID, rev string, err error) {
	// FIXME: Unimplemented
	return "", "", nil
}

func (d *db) PutContext(_ context.Context, docID string, doc interface{}) (rev string, err error) {
	// FIXME: Unimplemented
	return "", nil
}

func (d *db) DeleteContext(_ context.Context, docID, rev string) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", nil
}

func (d *db) InfoContext(_ context.Context) (*driver.DBInfo, error) {
	// FIXME: Unimplemented
	return nil, nil
}

func (d *db) CompactContext(_ context.Context) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) CompactViewContext(_ context.Context, _ string) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) ViewCleanupContext(_ context.Context) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) SecurityContext(_ context.Context) (*driver.Security, error) {
	// FIXME: Unimplemented
	return nil, nil
}
