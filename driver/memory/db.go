package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
)

var notYetImplemented = errors.Status(kivik.StatusNotImplemented, "kivik: not yet implemented in memory driver")

// database is an in-memory database representation.
type db struct {
	*client
	dbName string
	db     *database
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
	return nil, notYetImplemented
}

func (d *db) Query(ctx context.Context, ddoc, view string, opts map[string]interface{}) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) Get(_ context.Context, docID string, opts map[string]interface{}) (json.RawMessage, error) {
	if existing, ok := d.db.docs[docID]; ok {
		last := existing.revs[len(existing.revs)-1]
		return last.data, nil
	}
	return nil, errors.Status(kivik.StatusNotFound, "missing")
}

func (d *db) CreateDoc(_ context.Context, doc interface{}) (docID, rev string, err error) {
	// FIXME: Unimplemented
	return "", "", notYetImplemented
}

type docMeta struct {
	ID  string `json:"_id"`
	Rev string `json:"_rev"`
}

func (d *db) Put(_ context.Context, docID string, doc interface{}) (rev string, err error) {
	if existing, ok := d.db.docs[docID]; ok {
		last := existing.revs[len(existing.revs)-1]
		lastRev := fmt.Sprintf("%d-%s", last.RevID, last.Rev)
		if rev != lastRev {
			return "", errors.Status(kivik.StatusConflict, "document update conflict")
		}
	}
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return "", errors.Status(kivik.StatusBadRequest, "invalid JSON")
	}
	var meta docMeta
	if e := json.Unmarshal(docJSON, &meta); e != nil {
		return "", errors.Status(kivik.StatusInternalServerError, "failed to decode encoded document; this is a bug!")
	}
	if meta.Rev != "" {
		// Rev should not be set for a new document
		return "", errors.Status(kivik.StatusConflict, "document update conflict")
	}
	d.db.mutex.Lock()
	defer d.db.mutex.Unlock()
	revID := int64(1)
	revStr := randStr()
	d.db.docs[docID] = &document{
		revs: []*revision{
			{
				ID:    docID,
				RevID: revID,
				Rev:   revStr,
				data:  docJSON,
			},
		},
	}
	return fmt.Sprintf("%d-%s", revID, revStr), nil
}

func (d *db) Delete(_ context.Context, docID, rev string) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) Stats(_ context.Context) (*driver.DBStats, error) {
	return &driver.DBStats{
		Name: d.dbName,
		// DocCount:     0,
		// DeletedCount: 0,
		// UpdateSeq:    "",
		// DiskSize:     0,
		// ActiveSize:   0,
		// ExternalSize: 0,
	}, nil
}

func (c *client) Compact(_ context.Context) error {
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

func (d *db) Changes(ctx context.Context, opts map[string]interface{}) (driver.Changes, error) {
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
