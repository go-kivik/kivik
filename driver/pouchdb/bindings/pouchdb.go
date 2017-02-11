// Package bindings provides minimal GopherJS bindings around the PouchDB
// library. (https://pouchdb.com/api.html)
package bindings

import (
	"errors"

	"github.com/gopherjs/gopherjs/js"
	"github.com/gopherjs/jsbuiltin"
)

// DB is a PouchDB database object.
type DB struct {
	*js.Object
}

// New creates a database or opens an existing one.
//
// See https://pouchdb.com/api.html#create_database
func New(dbName string, options map[string]interface{}) *DB {
	return &DB{js.Global.Get("PouchDB").New(dbName, options)}
}

// Version returns the version of the currently running PouchDB library.
func Version() string {
	return js.Global.Get("PouchDB").Get("version").String()
}

// AllDBs returns the list of all existing (undeleted) databases.
func AllDBs() (allDBs []string, err error) {
	if jsbuiltin.TypeOf(js.Global.Get("PouchDB").Get("allDbs")) != "function" {
		return nil, errors.New("pouchdb-all-dbs plugin not loaded")
	}
	resultCh := make(chan *js.Object)
	js.Global.Get("PouchDB").Call("allDbs", func(e, r *js.Object) {
		if e != nil {
			err = &js.Error{Object: e}
		}
		resultCh <- r
	})
	result := <-resultCh
	if result != js.Undefined {
		for i := 0; i < result.Length(); i++ {
			allDBs = append(allDBs, result.Index(i).String())
		}
	}
	return allDBs, err
}
