package memory

import (
	"context"

	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/errors"
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

func (d *db) SetOption(_ string, _ interface{}) error {
	return errors.New("no options supported")
}

func (d *db) AllDocsContext(ctx context.Context, opts map[string]interface{}) (driver.Rows, error) {
	return nil, nil
	// if exists, _ := d.client.DBExistsContext(ctx, d.dbName); !exists {
	// 	return 0, 0, "", errors.Status(kivik.StatusNotFound, "database not found")
	// }
	// db := d.getDB()
	// db.mutex.RLock()
	// defer db.mutex.RUnlock()
	// index := make([]indexDoc, 0, len(db.docs))
	// for id, doc := range db.docs {
	// 	index = append(index, indexDoc{
	// 		ID:  id,
	// 		Key: id,
	// 		Value: indexDocValue{
	// 			Rev: doc.revs[len(doc.revs)-1].Rev,
	// 		},
	// 	})
	// }
	//
	// body, err := json.Marshal(ouchdb.AllDocsResponse{
	// 	Offset:    0,
	// 	TotalRows: len(index),
	// 	Rows:      index,
	// })
	// if err != nil {
	// 	panic(err)
	// }
	// return ouchdb.AllDocs(bytes.NewReader(body), docs)
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

func (c *client) CompactContext(_ context.Context) error {
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

func (d *db) SetSecurityContext(_ context.Context, _ *driver.Security) error {
	// FIXME: Unimplemented
	return nil
}

func (d *db) RevsLimitContext(_ context.Context) (limit int, err error) {
	// FIXME: Unimplemented
	return 0, nil
}

func (d *db) SetRevsLimitContext(_ context.Context, limit int) error {
	// FIXME: Unimplemented
	return nil
}
