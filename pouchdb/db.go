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

package pouchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sync/atomic"

	"github.com/gopherjs/gopherjs/js"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/pouchdb/bindings"
)

type db struct {
	db *bindings.DB

	client *client

	// these are set to 1 when compaction begins, and unset when the
	// callback returns.
	compacting  uint32
	viewCleanup uint32
}

var (
	_ driver.DB       = &db{}
	_ driver.DBCloser = &db{}
)

func (d *db) AllDocs(ctx context.Context, options map[interface{}]interface{}) (driver.Rows, error) {
	result, err := d.db.AllDocs(ctx, options)
	if err != nil {
		return nil, err
	}
	return &rows{
		Object: result,
	}, nil
}

func (d *db) Query(ctx context.Context, ddoc, view string, options map[interface{}]interface{}) (driver.Rows, error) {
	result, err := d.db.Query(ctx, ddoc, view, options)
	if err != nil {
		return nil, err
	}
	return &rows{
		Object: result,
	}, nil
}

func (d *db) Get(ctx context.Context, docID string, options map[interface{}]interface{}) (*driver.Document, error) {
	doc, rev, err := d.db.Get(ctx, docID, options)
	if err != nil {
		return nil, err
	}
	return &driver.Document{
		Rev:  rev,
		Body: ioutil.NopCloser(bytes.NewReader(doc)),
	}, nil
}

func (d *db) CreateDoc(ctx context.Context, doc interface{}, options map[interface{}]interface{}) (docID, rev string, err error) {
	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		return "", "", err
	}
	jsDoc := js.Global.Get("JSON").Call("parse", string(jsonDoc))
	return d.db.Post(ctx, jsDoc, options)
}

func (d *db) Put(ctx context.Context, docID string, doc interface{}, options map[interface{}]interface{}) (rev string, err error) {
	jsonDoc, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}
	jsDoc := js.Global.Get("JSON").Call("parse", string(jsonDoc))
	if id := jsDoc.Get("_id"); id != js.Undefined {
		if id.String() != docID {
			return "", &kivik.Error{Status: http.StatusBadRequest, Message: "id argument must match _id field in document"}
		}
	}
	jsDoc.Set("_id", docID)
	return d.db.Put(ctx, jsDoc, options)
}

func (d *db) Delete(ctx context.Context, docID string, options map[interface{}]interface{}) (newRev string, err error) {
	rev, _ := options["rev"].(string)
	return d.db.Delete(ctx, docID, rev, options)
}

func (d *db) Stats(ctx context.Context) (*driver.DBStats, error) {
	i, err := d.db.Info(ctx)
	return &driver.DBStats{
		Name:           i.Name,
		CompactRunning: atomic.LoadUint32(&d.compacting) == 1 || atomic.LoadUint32(&d.viewCleanup) == 1,
		DocCount:       i.DocCount,
		UpdateSeq:      i.UpdateSeq,
	}, err
}

func (d *db) Compact(_ context.Context) error {
	if atomic.LoadUint32(&d.compacting) == 1 {
		return &kivik.Error{Status: http.StatusTooManyRequests, Message: "kivik: compaction already running"}
	}
	atomic.StoreUint32(&d.compacting, 1)
	defer atomic.StoreUint32(&d.compacting, 0)
	return d.db.Compact()
}

// CompactView is unimplemented for PouchDB.
func (d *db) CompactView(context.Context, string) error {
	return nil
}

func (d *db) ViewCleanup(context.Context) error {
	if atomic.LoadUint32(&d.viewCleanup) == 1 {
		return &kivik.Error{Status: http.StatusTooManyRequests, Message: "kivik: view cleanup already running"}
	}
	atomic.StoreUint32(&d.viewCleanup, 1)
	defer atomic.StoreUint32(&d.viewCleanup, 0)
	return d.db.ViewCleanup()
}

func (d *db) Close() error {
	return d.db.Close()
}

var _ driver.Purger = &db{}

func (d *db) Purge(ctx context.Context, docRevMap map[string][]string) (*driver.PurgeResult, error) {
	result := &driver.PurgeResult{
		Purged: make(map[string][]string, len(docRevMap)),
	}
	for docID, revs := range docRevMap {
		for _, rev := range revs {
			delRevs, err := d.db.Purge(ctx, docID, rev)
			if err != nil {
				return result, err
			}
			if result.Purged[docID] == nil {
				result.Purged[docID] = delRevs
			} else {
				result.Purged[docID] = append(result.Purged[docID], delRevs...)
			}
		}
	}
	return result, nil
}
