package pouchdb

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/ouchdb"
	"github.com/flimzy/kivik/driver/pouchdb/bindings"
	"github.com/flimzy/kivik/errors"
	"github.com/gopherjs/gopherjs/js"
)

type db struct {
	db *bindings.DB
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

func (d *db) CreateDocContext(_ context.Context, doc interface{}) (docID, rev string, err error) {
	return "", "", nil
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
		Name:      i.Name,
		DocCount:  i.DocCount,
		UpdateSeq: i.UpdateSeq,
	}, err
}
