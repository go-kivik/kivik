package memorydb

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/mango"
)

var errFindNotImplemented = errors.New("find feature not yet implemented")

type findQuery struct {
	Selector *mango.Selector `json:"selector"`
	Limit    int64           `json:"limit"`
	Skip     int64           `json:"skip"`
	Sort     []string        `json:"sort"`
	Fields   []string        `json:"fields"`
	UseIndex indexSpec       `json:"use_index"`
}

type indexSpec struct {
	ddoc  string
	index string
}

func (i *indexSpec) UnmarshalJSON(data []byte) error {
	if data[0] == '"' {
		return json.Unmarshal(data, &i.ddoc)
	}
	var values []string
	if err := json.Unmarshal(data, &values); err != nil {
		return err
	}
	if len(values) == 0 || len(values) > 2 {
		return errors.New("invalid index specification")
	}
	i.ddoc = values[0]
	if len(values) == 2 {
		i.index = values[1]
	}
	return nil
}

var _ driver.Finder = &db{}

func (d *db) CreateIndex(context.Context, string, string, interface{}, driver.Options) error {
	return errFindNotImplemented
}

func (d *db) GetIndexes(context.Context, driver.Options) ([]driver.Index, error) {
	return nil, errFindNotImplemented
}

func (d *db) DeleteIndex(context.Context, string, string, driver.Options) error {
	return errFindNotImplemented
}

func (d *db) Find(ctx context.Context, query interface{}, _ driver.Options) (driver.Rows, error) {
	if exists, _ := d.DBExists(ctx, d.dbName, nil); !exists {
		return nil, statusError{status: http.StatusNotFound, error: errors.New("database does not exist")}
	}
	queryJSON, err := toJSON(query)
	if err != nil {
		return nil, err
	}
	fq := &findQuery{}
	if err := json.NewDecoder(queryJSON).Decode(&fq); err != nil {
		return nil, err
	}
	if fq == nil || fq.Selector == nil {
		return nil, errors.New("Missing required key: selector")
	}
	fields := make(map[string]struct{}, len(fq.Fields))
	for _, field := range fq.Fields {
		fields[field] = struct{}{}
	}
	rows := &findResults{
		resultSet: resultSet{
			docIDs: make([]string, 0),
			revs:   make([]*revision, 0),
		},
		fields: fields,
	}
	for docID := range d.db.docs {
		if doc, found := d.db.latestRevision(docID); found {
			var cd couchDoc
			if err := json.Unmarshal(doc.data, &cd); err != nil {
				panic(err)
			}
			match, err := fq.Selector.Matches(map[string]interface{}(cd))
			if err != nil {
				return nil, err
			}
			if match {
				rows.docIDs = append(rows.docIDs, docID)
				rows.revs = append(rows.revs, doc)
			}
		}
	}
	rows.offset = 0
	rows.totalRows = int64(len(rows.docIDs))
	return rows, nil
}

func (d *db) Explain(context.Context, interface{}, driver.Options) (*driver.QueryPlan, error) {
	return nil, errFindNotImplemented
}

type findResults struct {
	resultSet
	fields map[string]struct{}
}

var (
	_ driver.Rows       = &findResults{}
	_ driver.RowsWarner = &findResults{}
)

func (r *findResults) Warning() string {
	return "no matching index found, create an index to optimize query time"
}

func (r *findResults) Next(row *driver.Row) error {
	if r.revs == nil || len(r.revs) == 0 {
		return io.EOF
	}
	row.ID, r.docIDs = r.docIDs[0], r.docIDs[1:]
	doc, err := r.filterDoc(r.revs[0].data)
	if err != nil {
		return err
	}
	row.Doc = bytes.NewReader(doc)
	r.revs = r.revs[1:]
	return nil
}

func (r *findResults) filterDoc(data []byte) ([]byte, error) {
	if len(r.fields) == 0 {
		return data, nil
	}
	var intermediateDoc map[string]interface{}
	if err := json.Unmarshal(data, &intermediateDoc); err != nil {
		return nil, err
	}
	for field := range intermediateDoc {
		if _, ok := r.fields[field]; !ok {
			delete(intermediateDoc, field)
		}
	}
	return json.Marshal(intermediateDoc)
}

// toJSON converts a string, []byte, json.RawMessage, or an arbitrary type into
// an io.Reader of JSON marshaled data.
func toJSON(i interface{}) (io.Reader, error) {
	switch t := i.(type) {
	case string:
		return strings.NewReader(t), nil
	case []byte:
		return bytes.NewReader(t), nil
	case json.RawMessage:
		return bytes.NewReader(t), nil
	default:
		buf := &bytes.Buffer{}
		err := json.NewEncoder(buf).Encode(i)
		return buf, err
	}
}
