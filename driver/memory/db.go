package memory

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/flimzy/kivik/driver/ouchdb"
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

/* Options for AllDocs:

Per-doc options:
conflicts (boolean) – Includes conflicts information in response. Ignored if include_docs isn’t true. Default is false.
include_docs (boolean) – Include the full content of the documents in the return. Default is false.

Per-call options:
descending (boolean) – Return the documents in descending by key order. Default is false.
endkey (string) – Stop returning records when the specified key is reached. Optional.
end_key (string) – Alias for endkey param.
endkey_docid (string) – Stop returning records when the specified document ID is reached. Optional.
end_key_doc_id (string) – Alias for endkey_docid param.
inclusive_end (boolean) – Specifies whether the specified end key should be included in the result. Default is true.
key (string) – Return only documents that match the specified key. Optional.
keys (string) – Return only documents that match the specified keys. Optional.
limit (number) – Limit the number of the returned documents to the specified number. Optional.
skip (number) – Skip this number of records before starting to return the results. Default is 0.
stale (string) – Allow the results from a stale view to be used, without triggering a rebuild of all views within the encompassing design doc. Supported values: ok and update_after. Optional.
startkey (string) – Return records starting with the specified key. Optional.
start_key (string) – Alias for startkey param.
startkey_docid (string) – Return records starting with the specified document ID. Optional.
start_key_doc_id (string) – Alias for startkey_docid param.
update_seq (boolean) – Response includes an update_seq value indicating which sequence id of the underlying database the view reflects. Default is false.
*/

func (d *db) AllDocs(docs interface{}, opts map[string]interface{}) (offset, total int, seq string, err error) {
	if exists, _ := d.client.DBExists(d.dbName); !exists {
		return 0, 0, "", errors.Status(http.StatusNotFound, "database not found")
	}
	db := d.getDB()
	db.mutex.RLock()
	defer db.mutex.RUnlock()
	index := make([]indexDoc, 0, len(db.docs))
	for id, doc := range db.docs {
		index = append(index, indexDoc{
			ID:  id,
			Key: id,
			Value: indexDocValue{
				Rev: doc.revs[len(doc.revs)-1].Rev,
			},
		})
	}

	body, err := json.Marshal(ouchdb.AllDocsResponse{
		Offset:    0,
		TotalRows: len(index),
		Rows:      index,
	})
	if err != nil {
		panic(err)
	}
	return ouchdb.AllDocs(bytes.NewReader(body), docs)
}

func (d *db) Get(docID string, doc interface{}, opts map[string]interface{}) error {
	return nil
}

func (d *db) CreateDoc(doc interface{}) (docID, rev string, err error) {
	return "", "", nil
}

func (d *db) Put(docID string, doc interface{}) (rev string, err error) {
	return "", nil
}

func (d *db) Delete(docID, rev string) (newRev string, err error) {
	return "", nil
}
