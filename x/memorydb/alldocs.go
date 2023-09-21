package memorydb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func (d *db) AllDocs(ctx context.Context, _ driver.Options) (driver.Rows, error) {
	if exists, _ := d.DBExists(ctx, d.dbName, kivik.Params(nil)); !exists {
		return nil, statusError{status: http.StatusNotFound, error: errors.New("database does not exist")}
	}
	rows := &alldocsResults{
		resultSet{
			docIDs: make([]string, 0),
			revs:   make([]*revision, 0),
		},
	}
	for docID := range d.db.docs {
		if doc, found := d.db.latestRevision(docID); found {
			rows.docIDs = append(rows.docIDs, docID)
			rows.revs = append(rows.revs, doc)
		}
	}
	rows.offset = 0
	rows.totalRows = int64(len(rows.docIDs))
	return rows, nil
}

type resultSet struct {
	docIDs            []string
	revs              []*revision
	offset, totalRows int64
	updateSeq         string
}

func (r *resultSet) Close() error {
	r.revs = nil
	return nil
}

func (r *resultSet) UpdateSeq() string { return r.updateSeq }
func (r *resultSet) TotalRows() int64  { return r.totalRows }
func (r *resultSet) Offset() int64     { return r.offset }

type alldocsResults struct {
	resultSet
}

var _ driver.Rows = &alldocsResults{}

func (r *alldocsResults) Next(row *driver.Row) error {
	if r.revs == nil || len(r.revs) == 0 {
		return io.EOF
	}
	row.ID, r.docIDs = r.docIDs[0], r.docIDs[1:]
	var next *revision
	next, r.revs = r.revs[0], r.revs[1:]
	row.Key = []byte(fmt.Sprintf(`"%s"`, row.ID))
	value, err := json.Marshal(map[string]string{
		"rev": fmt.Sprintf("%d-%s", next.ID, next.Rev),
	})
	if err != nil {
		panic(err)
	}
	row.Value = bytes.NewReader(value)
	return nil
}
