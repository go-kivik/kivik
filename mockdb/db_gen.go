/* This file is auto-generated. Do not edit it! */

package kivikmock

import (
	"context"

	"github.com/go-kivik/kivik/v4/driver"
)

var _ = &driver.Attachment{}

func (db *driverDB) Compact(ctx context.Context) error {
	expected := &ExpectedCompact{
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx)
	}
	return expected.wait(ctx)
}

func (db *driverDB) CompactView(ctx context.Context, arg0 string) error {
	expected := &ExpectedCompactView{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0)
	}
	return expected.wait(ctx)
}

func (db *driverDB) Copy(ctx context.Context, arg0 string, arg1 string, options map[string]interface{}) (string, error) {
	expected := &ExpectedCopy{
		arg0: arg0,
		arg1: arg1,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) CreateDoc(ctx context.Context, arg0 interface{}, options map[string]interface{}) (string, string, error) {
	expected := &ExpectedCreateDoc{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return "", "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return expected.ret0, expected.ret1, expected.wait(ctx)
}

func (db *driverDB) CreateIndex(ctx context.Context, arg0 string, arg1 string, arg2 interface{}, options map[string]interface{}) error {
	expected := &ExpectedCreateIndex{
		arg0: arg0,
		arg1: arg1,
		arg2: arg2,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, arg2, options)
	}
	return expected.wait(ctx)
}

func (db *driverDB) DeleteIndex(ctx context.Context, arg0 string, arg1 string, options map[string]interface{}) error {
	expected := &ExpectedDeleteIndex{
		arg0: arg0,
		arg1: arg1,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, options)
	}
	return expected.wait(ctx)
}

func (db *driverDB) Flush(ctx context.Context) error {
	expected := &ExpectedFlush{
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx)
	}
	return expected.wait(ctx)
}

func (db *driverDB) GetRev(ctx context.Context, arg0 string, options map[string]interface{}) (string, error) {
	expected := &ExpectedGetRev{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) Put(ctx context.Context, arg0 string, arg1 interface{}, options map[string]interface{}) (string, error) {
	expected := &ExpectedPut{
		arg0: arg0,
		arg1: arg1,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) ViewCleanup(ctx context.Context) error {
	expected := &ExpectedViewCleanup{
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx)
	}
	return expected.wait(ctx)
}

func (db *driverDB) AllDocs(ctx context.Context, options map[string]interface{}) (driver.Rows, error) {
	expected := &ExpectedAllDocs{
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, options)
	}
	return &driverRows{Context: ctx, Rows: coalesceRows(expected.ret0)}, expected.wait(ctx)
}

func (db *driverDB) BulkDocs(ctx context.Context, arg0 []interface{}, options map[string]interface{}) ([]driver.BulkResult, error) {
	expected := &ExpectedBulkDocs{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) BulkGet(ctx context.Context, arg0 []driver.BulkGetReference, options map[string]interface{}) (driver.Rows, error) {
	expected := &ExpectedBulkGet{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return &driverRows{Context: ctx, Rows: coalesceRows(expected.ret0)}, expected.wait(ctx)
}

func (db *driverDB) Changes(ctx context.Context, options map[string]interface{}) (driver.Changes, error) {
	expected := &ExpectedChanges{
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, options)
	}
	return &driverChanges{Context: ctx, Changes: coalesceChanges(expected.ret0)}, expected.wait(ctx)
}

func (db *driverDB) Delete(ctx context.Context, arg0 string, options map[string]interface{}) (string, error) {
	expected := &ExpectedDelete{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) DeleteAttachment(ctx context.Context, arg0 string, arg1 string, options map[string]interface{}) (string, error) {
	expected := &ExpectedDeleteAttachment{
		arg0: arg0,
		arg1: arg1,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) DesignDocs(ctx context.Context, options map[string]interface{}) (driver.Rows, error) {
	expected := &ExpectedDesignDocs{
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, options)
	}
	return &driverRows{Context: ctx, Rows: coalesceRows(expected.ret0)}, expected.wait(ctx)
}

func (db *driverDB) Explain(ctx context.Context, arg0 interface{}, options map[string]interface{}) (*driver.QueryPlan, error) {
	expected := &ExpectedExplain{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) Find(ctx context.Context, arg0 interface{}, options map[string]interface{}) (driver.Rows, error) {
	expected := &ExpectedFind{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return &driverRows{Context: ctx, Rows: coalesceRows(expected.ret0)}, expected.wait(ctx)
}

func (db *driverDB) Get(ctx context.Context, arg0 string, options map[string]interface{}) (*driver.Document, error) {
	expected := &ExpectedGet{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) GetAttachment(ctx context.Context, arg0 string, arg1 string, options map[string]interface{}) (*driver.Attachment, error) {
	expected := &ExpectedGetAttachment{
		arg0: arg0,
		arg1: arg1,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) GetAttachmentMeta(ctx context.Context, arg0 string, arg1 string, options map[string]interface{}) (*driver.Attachment, error) {
	expected := &ExpectedGetAttachmentMeta{
		arg0: arg0,
		arg1: arg1,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) GetIndexes(ctx context.Context, options map[string]interface{}) ([]driver.Index, error) {
	expected := &ExpectedGetIndexes{
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) LocalDocs(ctx context.Context, options map[string]interface{}) (driver.Rows, error) {
	expected := &ExpectedLocalDocs{
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, options)
	}
	return &driverRows{Context: ctx, Rows: coalesceRows(expected.ret0)}, expected.wait(ctx)
}

func (db *driverDB) PartitionStats(ctx context.Context, arg0 string) (*driver.PartitionStats, error) {
	expected := &ExpectedPartitionStats{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) Purge(ctx context.Context, arg0 map[string][]string) (*driver.PurgeResult, error) {
	expected := &ExpectedPurge{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) PutAttachment(ctx context.Context, arg0 string, arg1 *driver.Attachment, options map[string]interface{}) (string, error) {
	expected := &ExpectedPutAttachment{
		arg0: arg0,
		arg1: arg1,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return "", err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, options)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) Query(ctx context.Context, arg0 string, arg1 string, options map[string]interface{}) (driver.Rows, error) {
	expected := &ExpectedQuery{
		arg0: arg0,
		arg1: arg1,
		commonExpectation: commonExpectation{
			db:      db.DB,
			options: options,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0, arg1, options)
	}
	return &driverRows{Context: ctx, Rows: coalesceRows(expected.ret0)}, expected.wait(ctx)
}

func (db *driverDB) RevsDiff(ctx context.Context, arg0 interface{}) (driver.Rows, error) {
	expected := &ExpectedRevsDiff{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0)
	}
	return &driverRows{Context: ctx, Rows: coalesceRows(expected.ret0)}, expected.wait(ctx)
}

func (db *driverDB) Security(ctx context.Context) (*driver.Security, error) {
	expected := &ExpectedSecurity{
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx)
	}
	return expected.ret0, expected.wait(ctx)
}

func (db *driverDB) SetSecurity(ctx context.Context, arg0 *driver.Security) error {
	expected := &ExpectedSetSecurity{
		arg0: arg0,
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return err
	}
	if expected.callback != nil {
		return expected.callback(ctx, arg0)
	}
	return expected.wait(ctx)
}

func (db *driverDB) Stats(ctx context.Context) (*driver.DBStats, error) {
	expected := &ExpectedStats{
		commonExpectation: commonExpectation{
			db: db.DB,
		},
	}
	if err := db.client.nextExpectation(expected); err != nil {
		return nil, err
	}
	if expected.callback != nil {
		return expected.callback(ctx)
	}
	return expected.ret0, expected.wait(ctx)
}
