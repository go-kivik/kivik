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

package memorydb

import (
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-kivik/kivik/v4/driver"
)

type file struct {
	ContentType string
	Data        []byte
}

type document struct {
	revs []*revision
}

type revision struct {
	data        []byte
	ID          int64
	Rev         string
	Deleted     bool
	Attachments map[string]file
}

type database struct {
	mu       sync.RWMutex
	docs     map[string]*document
	deleted  bool
	security *driver.Security
}

var (
	rnd   *rand.Rand
	rndMU = &sync.Mutex{}
)

func init() {
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
}

func (d *database) getRevision(docID, rev string) (*revision, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	doc, ok := d.docs[docID]
	if !ok {
		return nil, false
	}
	for _, r := range doc.revs {
		if rev == fmt.Sprintf("%d-%s", r.ID, r.Rev) {
			return r, true
		}
	}
	return nil, false
}

func (d *database) docExists(docID string) bool {
	d.mu.RLock()
	defer d.mu.RUnlock()
	_, ok := d.docs[docID]
	return ok
}

func (d *database) latestRevision(docID string) (*revision, bool) {
	d.mu.RLock()
	defer d.mu.RUnlock()
	doc, ok := d.docs[docID]
	if ok {
		last := doc.revs[len(doc.revs)-1]
		return last, true
	}
	return nil, false
}

type couchDoc map[string]interface{}

func (d couchDoc) ID() string {
	id, _ := d["_id"].(string)
	return id
}

func (d couchDoc) Rev() string {
	rev, _ := d["_rev"].(string)
	return rev
}

func toCouchDoc(i interface{}) (couchDoc, error) {
	if d, ok := i.(couchDoc); ok {
		return d, nil
	}
	asJSON, err := json.Marshal(i)
	if err != nil {
		return nil, statusError{status: http.StatusBadRequest, error: err}
	}
	var m couchDoc
	if e := json.Unmarshal(asJSON, &m); e != nil {
		return nil, statusError{status: http.StatusInternalServerError, error: errors.New("THIS IS A BUG: failed to decode encoded document")}
	}
	return m, nil
}

func (d *database) addRevision(doc couchDoc) string {
	d.mu.Lock()
	defer d.mu.Unlock()
	id, ok := doc["_id"].(string)
	if !ok {
		panic("_id missing or not a string")
	}
	isLocal := strings.HasPrefix(id, "_local/")
	if d.docs[id] == nil {
		d.docs[id] = &document{
			revs: make([]*revision, 0, 1),
		}
	}
	var revID int64
	var revStr string
	if isLocal {
		revID = 1
		revStr = "0"
	} else {
		l := len(d.docs[id].revs)
		if l == 0 {
			revID = 1
		} else {
			revID = d.docs[id].revs[l-1].ID + 1
		}
		revStr = randStr()
	}
	rev := fmt.Sprintf("%d-%s", revID, revStr)
	doc["_rev"] = rev
	data, err := json.Marshal(doc)
	if err != nil {
		panic(err)
	}
	deleted, _ := doc["_deleted"].(bool)
	newRev := &revision{
		data:    data,
		ID:      revID,
		Rev:     revStr,
		Deleted: deleted,
	}
	if isLocal {
		d.docs[id].revs = []*revision{newRev}
	} else {
		d.docs[id].revs = append(d.docs[id].revs, newRev)
	}
	return rev
}
