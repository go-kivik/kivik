package fs

import (
	"context"
	"encoding/json"
	"io"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

type db struct {
	*client
	dbName string
}

var notYetImplemented = errors.Status(kivik.StatusNotImplemented, "kivik: not yet implemented in fs driver")

func (d *db) AllDocs(_ context.Context, _ map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) Query(ctx context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) Get(_ context.Context, docID string, opts map[string]interface{}) (json.RawMessage, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) CreateDoc(_ context.Context, doc interface{}) (docID, rev string, err error) {
	// FIXME: Unimplemented
	return "", "", notYetImplemented
}

func (d *db) Put(_ context.Context, docID string, doc interface{}) (rev string, err error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) Delete(_ context.Context, docID, rev string) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) Stats(_ context.Context) (*driver.DBStats, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) Compact(_ context.Context) error {
	// FIXME: Unimplemented
	return notYetImplemented
}

func (d *db) CompactView(_ context.Context, _ string) error {
	// FIXME: Unimplemented
	return notYetImplemented
}

func (d *db) ViewCleanup(_ context.Context) error {
	// FIXME: Unimplemented
	return notYetImplemented
}

func (d *db) Security(_ context.Context) (*driver.Security, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) SetSecurity(_ context.Context, _ *driver.Security) error {
	// FIXME: Unimplemented
	return notYetImplemented
}

func (d *db) Changes(_ context.Context, _ map[string]interface{}) (driver.Changes, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) BulkDocs(_ context.Context, _ []interface{}) (driver.BulkResults, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) PutAttachment(_ context.Context, _, _, _, _ string, _ io.Reader) (string, error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) GetAttachment(ctx context.Context, docID, rev, filename string) (contentType string, md5sum driver.MD5sum, body io.ReadCloser, err error) {
	// FIXME: Unimplemented
	return "", driver.MD5sum{}, nil, notYetImplemented
}

func (d *db) DeleteAttachment(ctx context.Context, docID, rev, filename string) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}
