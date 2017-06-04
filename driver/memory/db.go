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
		if rev, ok := opts["rev"].(string); ok {
			for _, r := range existing.revs {
				if rev == fmt.Sprintf("%d-%s", r.ID, r.Rev) {
					return r.data, nil
				}
			}
			return nil, errors.Status(kivik.StatusNotFound, "missing")
		}
		last := existing.revs[len(existing.revs)-1]
		if last.Deleted {
			return nil, errors.Status(kivik.StatusNotFound, "missing")
		}
		return last.data, nil
	}
	return nil, errors.Status(kivik.StatusNotFound, "missing")
}

func (d *db) CreateDoc(_ context.Context, doc interface{}) (docID, rev string, err error) {
	// FIXME: Unimplemented
	return "", "", notYetImplemented
}

func (d *db) Put(_ context.Context, docID string, doc interface{}) (rev string, err error) {
	docJSON, err := json.Marshal(doc)
	if err != nil {
		return "", errors.Status(kivik.StatusBadRequest, "invalid JSON")
	}
	var couchDoc jsondoc
	if e := json.Unmarshal(docJSON, &couchDoc); e != nil {
		return "", errors.Status(kivik.StatusInternalServerError, "failed to decode encoded document; this is a bug!")
	}

	if last, ok := d.db.latestRevision(docID); ok {
		lastRev := fmt.Sprintf("%d-%s", last.ID, last.Rev)
		if couchDoc.Rev() != lastRev {
			return "", errors.Status(kivik.StatusConflict, "document update conflict")
		}
		rev := d.db.addRevision(docID, couchDoc)
		return rev, nil
	}

	if couchDoc.Rev() != "" {
		// Rev should not be set for a new document
		return "", errors.Status(kivik.StatusConflict, "document update conflict")
	}
	d.db.mutex.Lock()
	defer d.db.mutex.Unlock()
	return d.db.addRevision(docID, couchDoc), nil
}

func (d *db) Delete(ctx context.Context, docID, rev string) (newRev string, err error) {
	if _, ok := d.db.docs[docID]; !ok {
		return "", errors.Status(kivik.StatusNotFound, "missing")
	}
	return d.Put(ctx, docID, map[string]interface{}{
		"_id":      docID,
		"_rev":     rev,
		"_deleted": true,
	})
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
