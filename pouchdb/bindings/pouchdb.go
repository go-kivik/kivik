// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

//go:build js
// +build js

// Package bindings provides minimal GopherJS bindings around the PouchDB
// library. (https://pouchdb.com/api.html)
package bindings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"

	"github.com/go-kivik/kivik/v4/internal"
)

// DB is a PouchDB database object.
type DB struct {
	*js.Object
}

// PouchDB represents a PouchDB constructor.
type PouchDB struct {
	*js.Object
}

// GlobalPouchDB returns the global PouchDB object.
func GlobalPouchDB() *PouchDB {
	return &PouchDB{Object: js.Global.Get("PouchDB")}
}

// Defaults returns a new PouchDB constructor with the specified default options.
// See https://pouchdb.com/api.html#defaults
func Defaults(options map[string]interface{}) *PouchDB {
	return &PouchDB{Object: js.Global.Get("PouchDB").Call("defaults", options)}
}

// New creates a database or opens an existing one.
//
// See https://pouchdb.com/api.html#create_database
func (p *PouchDB) New(dbName string, options map[string]interface{}) *DB {
	db := &DB{Object: p.Object.New(dbName, options)}
	if db.indexeddb() {
		/* Without blocking here, we get the following error. This may be related
			to a sleep in PouchDB, that has a mysterious note about why it exists.
			https://github.com/pouchdb/pouchdb/blob/27ab3b27a6673038b449313d9700b3a7977ac091/packages/node_modules/pouchdb-adapter-indexeddb/src/index.js#L156-L160

		/home/jonhall/src/kivik/pouchdb/node_modules/pouchdb-adapter-indexeddb/lib/index.js:1597
			doc.rev_tree = pouchdbMerge.removeLeafFromTree(doc.rev_tree, rev);
																^
		TypeError: Cannot read properties of undefined (reading 'rev_tree')
			at FDBRequest.docStore.get.onsuccess (/home/jonhall/src/kivik/pouchdb/node_modules/pouchdb-adapter-indexeddb/lib/index.js:1597:58)
			at invokeEventListeners (/home/jonhall/src/kivik/pouchdb/node_modules/fake-indexeddb/build/cjs/lib/FakeEventTarget.js:55:25)
			at FDBRequest.dispatchEvent (/home/jonhall/src/kivik/pouchdb/node_modules/fake-indexeddb/build/cjs/lib/FakeEventTarget.js:99:7)
			at FDBTransaction._start (/home/jonhall/src/kivik/pouchdb/node_modules/fake-indexeddb/build/cjs/FDBTransaction.js:210:19)
			at Immediate.<anonymous> (/home/jonhall/src/kivik/pouchdb/node_modules/fake-indexeddb/build/cjs/lib/Database.js:38:16)
			at processImmediate (node:internal/timers:466:21)
		*/
		time.Sleep(0)
	}
	return db
}

// Version returns the version of the currently running PouchDB library.
func (p *PouchDB) Version() string {
	return p.Get("version").String()
}

func setTimeout(ctx context.Context, options map[string]interface{}) map[string]interface{} {
	if ctx == nil { // Just to be safe
		return options
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return options
	}
	if options == nil {
		options = make(map[string]interface{})
	}
	if _, ok := options["ajax"]; !ok {
		options["ajax"] = make(map[string]interface{})
	}
	ajax := options["ajax"].(map[string]interface{})
	timeout := int(time.Until(deadline) * 1000) //nolint:gomnd
	// Used by ajax calls
	ajax["timeout"] = timeout
	// Used by changes and replications
	options["timeout"] = timeout
	return options
}

type caller interface {
	Call(string, ...interface{}) *js.Object
}

// prepareArgs trims any trailing nil values, since JavaScript treats null as
// distinct from an omitted value.
func prepareArgs(args []interface{}) []interface{} {
	for len(args) > 0 {
		if !omitNil(args[len(args)-1]) {
			break
		}
		args = args[:len(args)-1]
	}
	return args
}

// omitNil returns true if a is a nil value that should be omitted as an
// argument to a JavaScript function.
func omitNil(a interface{}) bool {
	if a == nil {
		// a literal nil value should be converted to a null, so we don't omit
		return false
	}
	v := reflect.ValueOf(a)
	switch v.Kind() {
	case reflect.Slice, reflect.Interface, reflect.Map, reflect.Ptr:
		// nil slices, interfaces, maps, and pointers in our context mean that
		// we have a nil option that in JS idioms would just be omitted as an
		// argument, so return true.
		return v.IsNil()
	}
	return false
}

// callBack executes the 'method' of 'o' as a callback, setting result to the
// callback's return value. An error is returned if either the callback returns
// an error, or if the context is cancelled. No attempt is made to abort the
// callback in the case that the context is cancelled.
func callBack(ctx context.Context, o caller, method string, args ...interface{}) (r *js.Object, e error) {
	defer RecoverError(&e)
	resultCh := make(chan *js.Object)
	var err error
	o.Call(method, prepareArgs(args)...).Call("then", func(r *js.Object) {
		go func() { resultCh <- r }()
	}).Call("catch", func(e *js.Object) {
		err = NewPouchError(e)
		close(resultCh)
	})
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-resultCh:
		return result, err
	}
}

// AllDBs returns the list of all existing (undeleted) databases.
func (p *PouchDB) AllDBs(ctx context.Context) ([]string, error) {
	if jsbuiltin.TypeOf(p.Get("allDbs")) != jsbuiltin.TypeFunction {
		return nil, errors.New("pouchdb-all-dbs plugin not loaded")
	}
	result, err := callBack(ctx, p, "allDbs")
	if err != nil {
		return nil, err
	}
	if result == js.Undefined {
		return nil, nil
	}
	allDBs := make([]string, result.Length())
	for i := range allDBs {
		allDBs[i] = result.Index(i).String()
	}
	return allDBs, nil
}

// DBInfo is a struct respresenting information about a specific database.
type DBInfo struct {
	*js.Object
	Name      string `js:"db_name"`
	DocCount  int64  `js:"doc_count"`
	UpdateSeq string `js:"update_seq"`
}

// Info returns info about the database.
func (db *DB) Info(ctx context.Context) (*DBInfo, error) {
	result, err := callBack(ctx, db, "info")
	return &DBInfo{Object: result}, err
}

// Put creates a new document or update an existing document.
// See https://pouchdb.com/api.html#create_document
func (db *DB) Put(ctx context.Context, doc interface{}, opts map[string]interface{}) (rev string, err error) {
	result, err := callBack(ctx, db, "put", doc, setTimeout(ctx, opts))
	if err != nil {
		return "", err
	}
	return result.Get("rev").String(), nil
}

// Post creates a new document and lets PouchDB auto-generate the ID.
// See https://pouchdb.com/api.html#using-dbpost
func (db *DB) Post(ctx context.Context, doc interface{}, opts map[string]interface{}) (docID, rev string, err error) {
	result, err := callBack(ctx, db, "post", doc, setTimeout(ctx, opts))
	if err != nil {
		return "", "", err
	}
	return result.Get("id").String(), result.Get("rev").String(), nil
}

// Get fetches the requested document from the database.
// See https://pouchdb.com/api.html#fetch_document
func (db *DB) Get(ctx context.Context, docID string, opts map[string]interface{}) (doc []byte, rev string, err error) {
	result, err := callBack(ctx, db, "get", docID, setTimeout(ctx, opts))
	if err != nil {
		return nil, "", err
	}
	resultJSON := js.Global.Get("JSON").Call("stringify", result).String()
	return []byte(resultJSON), result.Get("_rev").String(), err
}

// Delete marks a document as deleted.
// See https://pouchdb.com/api.html#delete_document
func (db *DB) Delete(ctx context.Context, docID, rev string, opts map[string]interface{}) (newRev string, err error) {
	result, err := callBack(ctx, db, "remove", docID, rev, setTimeout(ctx, opts))
	if err != nil {
		return "", err
	}
	return result.Get("rev").String(), nil
}

func (db *DB) indexeddb() bool {
	return db.Object.Get("__opts").Get("adapter").String() == "indexeddb"
}

// Purge purges a specific document revision. It returns a list of successfully
// purged revisions. This method is only supported by the IndexedDB adaptor, and
// all others return an error.
func (db *DB) Purge(ctx context.Context, docID, rev string) ([]string, error) {
	if db.Object.Get("purge") == js.Undefined {
		return nil, &internal.Error{Status: http.StatusNotImplemented, Message: "kivik: purge supported by PouchDB 8 or newer"}
	}
	if !db.indexeddb() {
		return nil, &internal.Error{Status: http.StatusNotImplemented, Message: "kivik: purge only supported with indexedDB adapter"}
	}
	result, err := callBack(ctx, db, "purge", docID, rev, setTimeout(ctx, nil))
	if err != nil {
		return nil, err
	}
	delRevs := result.Get("deletedRevs")
	revs := make([]string, delRevs.Length())
	for i := range revs {
		revs[i] = delRevs.Index(i).String()
	}
	return revs, nil
}

// Destroy destroys the database.
func (db *DB) Destroy(ctx context.Context, options map[string]interface{}) error {
	_, err := callBack(ctx, db, "destroy", setTimeout(ctx, options))
	return err
}

// AllDocs returns a list of all documents in the database.
func (db *DB) AllDocs(ctx context.Context, options map[string]interface{}) (*js.Object, error) {
	return callBack(ctx, db, "allDocs", setTimeout(ctx, options))
}

// Query queries a map/reduce function.
func (db *DB) Query(ctx context.Context, ddoc, view string, options map[string]interface{}) (*js.Object, error) {
	o := setTimeout(ctx, options)
	return callBack(ctx, db, "query", ddoc+"/"+view, o)
}

const findPluginNotLoaded = internal.CompositeError("\x65pouchdb-find plugin not loaded")

// Find executes a MongoDB-style find query with the pouchdb-find plugin, if it
// is installed. If the plugin is not installed, a NotImplemented error will be
// returned.
//
// See https://github.com/pouchdb/pouchdb/tree/master/packages/node_modules/pouchdb-find#dbfindrequest--callback
func (db *DB) Find(ctx context.Context, query interface{}) (*js.Object, error) {
	if jsbuiltin.TypeOf(db.Object.Get("find")) != jsbuiltin.TypeFunction {
		return nil, findPluginNotLoaded
	}
	queryObj, err := Objectify(query)
	if err != nil {
		return nil, err
	}
	return callBack(ctx, db, "find", queryObj)
}

// Objectify unmarshals a string, []byte, or json.RawMessage into an interface{}.
// All other types are just passed through.
func Objectify(i interface{}) (interface{}, error) {
	var buf []byte
	switch t := i.(type) {
	case string:
		buf = []byte(t)
	case []byte:
		buf = t
	case json.RawMessage:
		buf = t
	default:
		return i, nil
	}
	var x interface{}
	err := json.Unmarshal(buf, &x)
	if err != nil {
		err = &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	return x, err
}

// Compact compacts the database, and waits for it to complete. This may take
// a long time! Please wrap this call in a goroutine.
func (db *DB) Compact() error {
	_, err := callBack(context.Background(), db, "compact")
	return err
}

// ViewCleanup cleans up views, and waits for it to complete. This may take a
// long time! Please wrap this call in a goroutine.
func (db *DB) ViewCleanup() error {
	_, err := callBack(context.Background(), db, "viewCleanup")
	return err
}

var jsJSON = js.Global.Get("JSON")

// BulkDocs creates, updates, or deletes docs in bulk.
// See https://pouchdb.com/api.html#batch_create
func (db *DB) BulkDocs(ctx context.Context, docs []interface{}, options map[string]interface{}) (result *js.Object, err error) {
	defer RecoverError(&err)
	jsDocs := make([]*js.Object, len(docs))
	for i, doc := range docs {
		jsonDoc, err := json.Marshal(doc)
		if err != nil {
			return nil, err
		}
		jsDocs[i] = jsJSON.Call("parse", string(jsonDoc))
	}
	if options == nil {
		return callBack(ctx, db, "bulkDocs", jsDocs, setTimeout(ctx, nil))
	}
	return callBack(ctx, db, "bulkDocs", jsDocs, options, setTimeout(ctx, nil))
}

// Changes returns an event emitter object.
//
// See https://pouchdb.com/api.html#changes
func (db *DB) Changes(ctx context.Context, options map[string]interface{}) (changes *js.Object, e error) {
	defer RecoverError(&e)
	return db.Call("changes", setTimeout(ctx, options)), nil
}

// PutAttachment attaches a binary object to a document.
//
// See https://pouchdb.com/api.html#save_attachment
func (db *DB) PutAttachment(ctx context.Context, docID, filename, rev string, body io.Reader, ctype string) (*js.Object, error) {
	att, err := attachmentObject(ctype, body)
	if err != nil {
		return nil, err
	}
	if rev == "" {
		return callBack(ctx, db, "putAttachment", docID, filename, att, ctype)
	}
	return callBack(ctx, db, "putAttachment", docID, filename, rev, att, ctype)
}

// attachmentObject converts an io.Reader to a JavaScript Buffer in node, or
// a Blob in the browser
func attachmentObject(contentType string, content io.Reader) (att *js.Object, err error) {
	RecoverError(&err)
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(content); err != nil {
		return nil, err
	}
	if buffer := js.Global.Get("Buffer"); jsbuiltin.TypeOf(buffer) == jsbuiltin.TypeFunction {
		// The Buffer type is supported, so we'll use that
		if jsbuiltin.TypeOf(buffer.Get("from")) == jsbuiltin.TypeFunction {
			// For newer versions of Node.js. See https://nodejs.org/fa/docs/guides/buffer-constructor-deprecation/
			return buffer.Call("from", buf.String()), nil
		}
		// Fall back to legacy Buffer constructor.
		return buffer.New(buf.String()), nil
	}
	if js.Global.Get("Blob") != js.Undefined {
		// We have Blob support, must be in a browser
		return js.Global.Get("Blob").New([]interface{}{buf.Bytes()}, map[string]string{"type": contentType}), nil
	}
	// Not sure what to do
	return nil, errors.New("No Blob or Buffer support?!?")
}

// GetAttachment returns attachment data.
//
// See https://pouchdb.com/api.html#get_attachment
func (db *DB) GetAttachment(ctx context.Context, docID, filename string, options map[string]interface{}) (*js.Object, error) {
	return callBack(ctx, db, "getAttachment", docID, filename, setTimeout(ctx, options))
}

// RemoveAttachment deletes an attachment from a document.
//
// See https://pouchdb.com/api.html#delete_attachment
func (db *DB) RemoveAttachment(ctx context.Context, docID, filename, rev string) (*js.Object, error) {
	return callBack(ctx, db, "removeAttachment", docID, filename, rev)
}

// CreateIndex creates an index to be used by MongoDB-style queries with the
// pouchdb-find plugin, if it is installed. If the plugin is not installed, a
// NotImplemented error will be returned.
//
// See https://github.com/pouchdb/pouchdb/tree/master/packages/node_modules/pouchdb-find#dbcreateindexindex--callback
func (db *DB) CreateIndex(ctx context.Context, index interface{}) (*js.Object, error) {
	if jsbuiltin.TypeOf(db.Object.Get("find")) != jsbuiltin.TypeFunction {
		return nil, findPluginNotLoaded
	}
	return callBack(ctx, db, "createIndex", index)
}

// GetIndexes returns the list of currently defined indexes on the database.
//
// See https://github.com/pouchdb/pouchdb/tree/master/packages/node_modules/pouchdb-find#dbgetindexescallback
func (db *DB) GetIndexes(ctx context.Context) (*js.Object, error) {
	if jsbuiltin.TypeOf(db.Object.Get("find")) != jsbuiltin.TypeFunction {
		return nil, findPluginNotLoaded
	}
	return callBack(ctx, db, "getIndexes")
}

// DeleteIndex deletes an index used by the MongoDB-style queries with the
// pouchdb-find plugin, if it is installed. If the plugin is not installed, a
// NotImplemeneted error will be returned.
//
// See: https://github.com/pouchdb/pouchdb/tree/master/packages/node_modules/pouchdb-find#dbdeleteindexindex--callback
func (db *DB) DeleteIndex(ctx context.Context, index interface{}) (*js.Object, error) {
	if jsbuiltin.TypeOf(db.Object.Get("find")) != jsbuiltin.TypeFunction {
		return nil, findPluginNotLoaded
	}
	return callBack(ctx, db, "deleteIndex", index)
}

// Replication events
const (
	ReplicationEventChange   = "change"
	ReplicationEventComplete = "complete"
	ReplicationEventPaused   = "paused"
	ReplicationEventActive   = "active"
	ReplicationEventDenied   = "denied"
	ReplicationEventError    = "error"
)

// Replicate initiates a replication.
// See https://pouchdb.com/api.html#replication
func (p *PouchDB) Replicate(source, target interface{}, options map[string]interface{}) (result *js.Object, err error) {
	defer RecoverError(&err)
	return p.Call("replicate", source, target, options), nil
}

// Explain the query plan for a given query
//
// See https://pouchdb.com/api.html#explain_index
func (db *DB) Explain(ctx context.Context, query interface{}) (*js.Object, error) {
	if jsbuiltin.TypeOf(db.Object.Get("find")) != jsbuiltin.TypeFunction {
		return nil, findPluginNotLoaded
	}
	queryObj, err := Objectify(query)
	if err != nil {
		return nil, err
	}
	return callBack(ctx, db, "explain", queryObj)
}

// Close closes the underlying db object.
func (db *DB) Close() error {
	// I'm not sure when DB.close() was added to PouchDB, so guard against
	// it missing, just in case.
	if jsbuiltin.TypeOf(db.Object.Get("close")) != jsbuiltin.TypeFunction {
		return nil
	}
	_, err := callBack(context.Background(), db, "close")
	return err
}
