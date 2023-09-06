/* This file is auto-generated. Do not edit it! */

package kivikmock

import (
	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

var _ = kivik.EndKeySuffix // To ensure a reference to kivik package
var _ = &driver.Attachment{}

// ExpectCompact queues an expectation that DB.Compact will be called.
func (db *DB) ExpectCompact() *ExpectedCompact {
	e := &ExpectedCompact{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectCompactView queues an expectation that DB.CompactView will be called.
func (db *DB) ExpectCompactView() *ExpectedCompactView {
	e := &ExpectedCompactView{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectCopy queues an expectation that DB.Copy will be called.
func (db *DB) ExpectCopy() *ExpectedCopy {
	e := &ExpectedCopy{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectCreateDoc queues an expectation that DB.CreateDoc will be called.
func (db *DB) ExpectCreateDoc() *ExpectedCreateDoc {
	e := &ExpectedCreateDoc{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectCreateIndex queues an expectation that DB.CreateIndex will be called.
func (db *DB) ExpectCreateIndex() *ExpectedCreateIndex {
	e := &ExpectedCreateIndex{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectDeleteIndex queues an expectation that DB.DeleteIndex will be called.
func (db *DB) ExpectDeleteIndex() *ExpectedDeleteIndex {
	e := &ExpectedDeleteIndex{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectFlush queues an expectation that DB.Flush will be called.
func (db *DB) ExpectFlush() *ExpectedFlush {
	e := &ExpectedFlush{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectGetRev queues an expectation that DB.GetRev will be called.
func (db *DB) ExpectGetRev() *ExpectedGetRev {
	e := &ExpectedGetRev{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectPut queues an expectation that DB.Put will be called.
func (db *DB) ExpectPut() *ExpectedPut {
	e := &ExpectedPut{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectViewCleanup queues an expectation that DB.ViewCleanup will be called.
func (db *DB) ExpectViewCleanup() *ExpectedViewCleanup {
	e := &ExpectedViewCleanup{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectAllDocs queues an expectation that DB.AllDocs will be called.
func (db *DB) ExpectAllDocs() *ExpectedAllDocs {
	e := &ExpectedAllDocs{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectBulkDocs queues an expectation that DB.BulkDocs will be called.
func (db *DB) ExpectBulkDocs() *ExpectedBulkDocs {
	e := &ExpectedBulkDocs{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectBulkGet queues an expectation that DB.BulkGet will be called.
func (db *DB) ExpectBulkGet() *ExpectedBulkGet {
	e := &ExpectedBulkGet{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectChanges queues an expectation that DB.Changes will be called.
func (db *DB) ExpectChanges() *ExpectedChanges {
	e := &ExpectedChanges{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectDelete queues an expectation that DB.Delete will be called.
func (db *DB) ExpectDelete() *ExpectedDelete {
	e := &ExpectedDelete{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectDeleteAttachment queues an expectation that DB.DeleteAttachment will be called.
func (db *DB) ExpectDeleteAttachment() *ExpectedDeleteAttachment {
	e := &ExpectedDeleteAttachment{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectDesignDocs queues an expectation that DB.DesignDocs will be called.
func (db *DB) ExpectDesignDocs() *ExpectedDesignDocs {
	e := &ExpectedDesignDocs{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectExplain queues an expectation that DB.Explain will be called.
func (db *DB) ExpectExplain() *ExpectedExplain {
	e := &ExpectedExplain{
		commonExpectation: commonExpectation{db: db},
		ret0:              &driver.QueryPlan{},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectFind queues an expectation that DB.Find will be called.
func (db *DB) ExpectFind() *ExpectedFind {
	e := &ExpectedFind{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectGet queues an expectation that DB.Get will be called.
func (db *DB) ExpectGet() *ExpectedGet {
	e := &ExpectedGet{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectGetAttachment queues an expectation that DB.GetAttachment will be called.
func (db *DB) ExpectGetAttachment() *ExpectedGetAttachment {
	e := &ExpectedGetAttachment{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectGetAttachmentMeta queues an expectation that DB.GetAttachmentMeta will be called.
func (db *DB) ExpectGetAttachmentMeta() *ExpectedGetAttachmentMeta {
	e := &ExpectedGetAttachmentMeta{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectGetIndexes queues an expectation that DB.GetIndexes will be called.
func (db *DB) ExpectGetIndexes() *ExpectedGetIndexes {
	e := &ExpectedGetIndexes{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectLocalDocs queues an expectation that DB.LocalDocs will be called.
func (db *DB) ExpectLocalDocs() *ExpectedLocalDocs {
	e := &ExpectedLocalDocs{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectPartitionStats queues an expectation that DB.PartitionStats will be called.
func (db *DB) ExpectPartitionStats() *ExpectedPartitionStats {
	e := &ExpectedPartitionStats{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectPurge queues an expectation that DB.Purge will be called.
func (db *DB) ExpectPurge() *ExpectedPurge {
	e := &ExpectedPurge{
		commonExpectation: commonExpectation{db: db},
		ret0:              &driver.PurgeResult{},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectPutAttachment queues an expectation that DB.PutAttachment will be called.
func (db *DB) ExpectPutAttachment() *ExpectedPutAttachment {
	e := &ExpectedPutAttachment{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectQuery queues an expectation that DB.Query will be called.
func (db *DB) ExpectQuery() *ExpectedQuery {
	e := &ExpectedQuery{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectRevsDiff queues an expectation that DB.RevsDiff will be called.
func (db *DB) ExpectRevsDiff() *ExpectedRevsDiff {
	e := &ExpectedRevsDiff{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectSecurity queues an expectation that DB.Security will be called.
func (db *DB) ExpectSecurity() *ExpectedSecurity {
	e := &ExpectedSecurity{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectSetSecurity queues an expectation that DB.SetSecurity will be called.
func (db *DB) ExpectSetSecurity() *ExpectedSetSecurity {
	e := &ExpectedSetSecurity{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}

// ExpectStats queues an expectation that DB.Stats will be called.
func (db *DB) ExpectStats() *ExpectedStats {
	e := &ExpectedStats{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}
