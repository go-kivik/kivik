package pouchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/ouchdb"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
	"github.com/flimzy/kivik/errors"
	"github.com/gopherjs/gopherjs/js"
	"github.com/imdario/mergo"
)

type db struct {
	db *bindings.DB

	// compacting is set true when compaction begins, and unset when the
	// callback returns.
	compacting bool

	opts map[string]interface{}
}

func (d *db) options(opts map[string]interface{}) (map[string]interface{}, error) {
	o := Options{}
	if err := mergo.MergeWithOverwrite(&o, d.opts); err != nil {
		return nil, err
	}
	return o, mergo.MergeWithOverwrite(&o, opts)
}

func (d *db) SetOption(key string, value interface{}) error {
	d.opts[key] = value
	return nil
}

func (d *db) AllDocsContext(ctx context.Context, docs interface{}, options map[string]interface{}) (offset, totalrows int, updateSeq string, err error) {
	body, err := d.db.AllDocs(ctx, options)
	if err != nil {
		return 0, 0, "", err
	}
	return ouchdb.AllDocs(bytes.NewReader(body), docs)
}

func (d *db) GetContext(ctx context.Context, docID string, doc interface{}, options map[string]interface{}) error {
	body, err := d.db.Get(ctx, docID, options)
	if err != nil {
		return err
	}
	return json.Unmarshal(body, &doc)
}

func (d *db) CreateDocContext(ctx context.Context, doc interface{}) (docID, rev string, err error) {
	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		return "", "", err
	}
	jsDoc := js.Global.Get("JSON").Call("parse", string(jsonDoc))
	return d.db.Post(ctx, jsDoc)
}

func (d *db) PutContext(ctx context.Context, docID string, doc interface{}) (rev string, err error) {
	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	jsDoc := js.Global.Get("JSON").Call("parse", string(jsonDoc))
	if id := jsDoc.Get("_id"); id != js.Undefined {
		if id.String() != docID {
			return "", errors.Status(kivik.StatusBadRequest, "id argument must match _id field in document")
		}
	}
	jsDoc.Set("_id", docID)
	return d.db.Put(ctx, jsDoc)
}

func (d *db) DeleteContext(ctx context.Context, docID, rev string) (newRev string, err error) {
	return d.db.Delete(ctx, map[string]string{
		"_id":  docID,
		"_rev": rev,
	})
}

func (d *db) InfoContext(ctx context.Context) (*driver.DBInfo, error) {
	i, err := d.db.Info(ctx)
	return &driver.DBInfo{
		Name:           i.Name,
		CompactRunning: d.compacting,
		DocCount:       i.DocCount,
		UpdateSeq:      i.UpdateSeq,
	}, err
}

func (d *db) CompactContext(_ context.Context) error {
	d.compacting = true
	go func() {
		defer func() { d.compacting = false }()
		if err := d.db.Compact(); err != nil {
			fmt.Fprintf(os.Stderr, "compaction failed: %s", err)
		}
	}()
	return nil
}

// CompactViewContext is unimplemented for PouchDB
func (d *db) CompactViewContext(_ context.Context, _ string) error {
	return nil
}

func (d *db) ViewCleanupContext(_ context.Context) error {
	d.compacting = true
	go func() {
		defer func() { d.compacting = false }()
		if err := d.db.ViewCleanup(); err != nil {
			fmt.Fprintf(os.Stderr, "view cleanup failed: %s", err)
		}
	}()
	return nil
}

func (d *db) SecurityContext(ctx context.Context) (*driver.Security, error) {
	return nil, kivik.ErrNotImplemented
}

func (d *db) SetSecurityContext(_ context.Context, _ *driver.Security) error {
	return kivik.ErrNotImplemented
}

func (d *db) RevsLimitContext(_ context.Context) (limit int, err error) {
	return 0, nil
}

func (d *db) SetRevsLimitContext(_ context.Context, limit int) error {
	return nil
}
