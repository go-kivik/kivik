package memory

import (
	"context"
	"encoding/json"
	"io"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
)

// database is an in-memory database representation.
type db struct {
	*client
	dbName string
}

type indexDoc struct {
	ID    string        `json:"id"`
	Key   string        `json:"key"`
	Value indexDocValue `json:"value"`
}

type indexDocValue struct {
	Rev string `json:"rev"`
}

func (d *db) AllDocs(ctx context.Context, opts map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, kivik.ErrNotImplemented
}

func (d *db) Query(ctx context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, kivik.ErrNotImplemented
}

func (d *db) Get(_ context.Context, docID string, opts map[string]interface{}) (json.RawMessage, error) {
	// FIXME: Unimplemented
	return nil, kivik.ErrNotImplemented
}

func (d *db) CreateDoc(_ context.Context, doc interface{}) (docID, rev string, err error) {
	// FIXME: Unimplemented
	return "", "", kivik.ErrNotImplemented
}

func (d *db) Put(_ context.Context, docID string, doc interface{}) (rev string, err error) {
	// FIXME: Unimplemented
	return "", kivik.ErrNotImplemented
}

func (d *db) Delete(_ context.Context, docID, rev string) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", kivik.ErrNotImplemented
}

func (d *db) Stats(_ context.Context) (*driver.DBStats, error) {
	// FIXME: Unimplemented
	return nil, kivik.ErrNotImplemented
}

func (c *client) Compact(_ context.Context) error {
	// FIXME: Unimplemented
	return kivik.ErrNotImplemented
}

func (d *db) CompactView(_ context.Context, _ string) error {
	// FIXME: Unimplemented
	return kivik.ErrNotImplemented
}

func (d *db) ViewCleanup(_ context.Context) error {
	// FIXME: Unimplemented
	return kivik.ErrNotImplemented
}

func (d *db) Security(_ context.Context) (*driver.Security, error) {
	// FIXME: Unimplemented
	return nil, kivik.ErrNotImplemented
}

func (d *db) SetSecurity(_ context.Context, _ *driver.Security) error {
	// FIXME: Unimplemented
	return kivik.ErrNotImplemented
}

func (d *db) Changes(ctx context.Context, opts map[string]interface{}) (driver.Changes, error) {
	// FIXME: Unimplemented
	return nil, kivik.ErrNotImplemented
}

func (d *db) BulkDocs(_ context.Context, _ ...interface{}) (driver.BulkResults, error) {
	// FIXME: Unimplemented
	return nil, kivik.ErrNotImplemented
}

func (d *db) PutAttachment(_ context.Context, _, _, _, _ string, _ io.Reader) (string, error) {
	// FIXME: Unimplemented
	return "", kivik.ErrNotImplemented
}

func (d *db) GetAttachment(ctx context.Context, docID, rev, filename string) (contentType string, md5sum driver.Checksum, body io.ReadCloser, err error) {
	// FIXME: Unimplemented
	return "", driver.Checksum{}, nil, kivik.ErrNotImplemented
}

func (d *db) DeleteAttachment(ctx context.Context, docID, rev, filename string) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", kivik.ErrNotImplemented
}
