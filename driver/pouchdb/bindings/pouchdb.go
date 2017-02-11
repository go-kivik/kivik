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
	return p.New(dbName, options)
}

// Version returns the version of the currently running PouchDB library.
func (p *PouchDB) Version() string {
	return p.Get("version").String()
}

// AllDBs returns the list of all existing (undeleted) databases.
func (p *PouchDB) AllDBs() ([]string, error) {
	if jsbuiltin.TypeOf(p.Get("allDbs")) != "function" {
		return nil, errors.New("pouchdb-all-dbs plugin not loaded")
	}
	resultCh := make(chan *js.Object)
	var err error
	p.Call("allDbs", func(e, r *js.Object) {
		if e != nil {
			err = &js.Error{Object: e}
		}
		resultCh <- r
	})
	result := <-resultCh
	var allDBs []string
	if result != js.Undefined {
		for i := 0; i < result.Length(); i++ {
			allDBs = append(allDBs, result.Index(i).String())
		}
	}
	return allDBs, err
}

// DBInfo is a struct respresenting information about a specific database.
type DBInfo struct {
	*js.Object
	DBName    string `js:"db_name"`
	DocCount  int    `js:"doc_count"`
	UpdateSeq string `js:"update_seq"`
}

// Info returns info about the database.
func (db *DB) Info() (*DBInfo, error) {
	resultCh := make(chan *DBInfo)
	var err error
	db.Call("info", func(e *js.Object, i *DBInfo) {
		if e != nil {
			err = &js.Error{Object: e}
		}
		resultCh <- i
	})
	info := <-resultCh
	return info, err
}
