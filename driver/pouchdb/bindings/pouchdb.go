// Package bindings provides minimal GopherJS bindings around the PouchDB
// library. (https://pouchdb.com/api.html)
package bindings

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"

	"github.com/flimzy/kivik/errors"
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
	return &DB{Object: p.Object.New(dbName, options)}
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
	ajax["timeout"] = int(deadline.Sub(time.Now()) * 1000)
	return options
}

type caller interface {
	Call(string, ...interface{}) *js.Object
}

// callBack executes the 'method' of 'o' as a callback, setting result to the
// callback's return value. An error is returned if either the callback returns
// an error, or if the context is cancelled. No attempt is made to abort the
// callback in the case that the context is cancelled.
func callBack(ctx context.Context, o caller, method string, args ...interface{}) (r *js.Object, e error) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case *js.Object:
				e = newPouchError(r.(*js.Object))
			case error:
				// This shouldn't ever happen, but just in case
				e = r.(error)
			default:
				// Catch all for everything else
				e = fmt.Errorf("%v", r)
			}
		}
	}()
	resultCh := make(chan *js.Object)
	var err error
	o.Call(method, args...).Call("then", func(r *js.Object) {
		resultCh <- r
	}).Call("catch", func(e *js.Object) {
		err = newPouchError(e)
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
	if jsbuiltin.TypeOf(p.Get("allDbs")) != "function" {
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
func (db *DB) Put(ctx context.Context, doc interface{}) (rev string, err error) {
	result, err := callBack(ctx, db, "put", doc, setTimeout(ctx, nil))
	if err != nil {
		return "", err
	}
	return result.Get("rev").String(), nil
}

// Post creates a new document and lets PouchDB auto-generate the ID.
// See https://pouchdb.com/api.html#using-dbpost
func (db *DB) Post(ctx context.Context, doc interface{}) (docID, rev string, err error) {
	result, err := callBack(ctx, db, "post", doc, setTimeout(ctx, nil))
	if err != nil {
		return "", "", err
	}
	return result.Get("id").String(), result.Get("rev").String(), nil
}

// Get fetches the requested document from the database.
// See https://pouchdb.com/api.html#fetch_document
func (db *DB) Get(ctx context.Context, docID string, opts map[string]interface{}) (doc []byte, err error) {
	result, err := callBack(ctx, db, "get", docID, setTimeout(ctx, opts))
	if err != nil {
		return nil, err
	}
	resultJSON := js.Global.Get("JSON").Call("stringify", result).String()
	return []byte(resultJSON), err
}

// Delete marks a document as deleted.
// See https://pouchdb.com/api.html#delete_document
func (db *DB) Delete(ctx context.Context, doc interface{}) (rev string, err error) {
	result, err := callBack(ctx, db, "remove", doc, setTimeout(ctx, nil))
	if err != nil {
		return "", err
	}
	return result.Get("rev").String(), nil
}

// Destroy destroys the database.
func (db *DB) Destroy(ctx context.Context, options map[string]interface{}) error {
	_, err := callBack(ctx, db, "destroy", setTimeout(ctx, options))
	return err
}

// AllDocs returns a list of all documents in the database.
func (db *DB) AllDocs(ctx context.Context, options map[string]interface{}) ([]byte, error) {
	result, err := callBack(ctx, db, "allDocs", setTimeout(ctx, options))
	if err != nil {
		return nil, err
	}
	resultJSON := js.Global.Get("JSON").Call("stringify", result).String()
	return []byte(resultJSON), err
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

const defaultRevsLimit = 1000

// RevsLimit returns the current revs_limit setting for the database.
func (db *DB) RevsLimit() (limit int, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	if db.Object.Get("_adaptor").String() == "http" {
		return 0, errors.Status(http.StatusNotImplemented, "revs_limit unimplemented for remote databases")
	}
	if revsLimit := db.Object.Get("__opts").Get("revs_limit"); revsLimit != js.Undefined {
		return revsLimit.Int(), nil
	}
	return defaultRevsLimit, nil
}
