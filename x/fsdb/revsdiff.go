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

package fs

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

var _ driver.RevsDiffer = &db{}

func toRevmap(i any) (map[string][]string, error) {
	if t, ok := i.(map[string][]string); ok {
		return t, nil
	}
	encoded, err := json.Marshal(i)
	if err != nil {
		return nil, statusError{status: http.StatusBadRequest, error: err}
	}
	revmap := make(map[string][]string)
	err = json.Unmarshal(encoded, &revmap)
	return revmap, statusError{status: http.StatusBadRequest, error: err}
}

func (d *db) RevsDiff(ctx context.Context, revMap any) (driver.Rows, error) {
	revmap, err := toRevmap(revMap)
	if err != nil {
		return nil, err
	}
	return &revDiffRows{
		ctx:    ctx,
		db:     d,
		revmap: revmap,
	}, nil
}

type revDiffRows struct {
	ctx    context.Context
	db     *db
	revmap map[string][]string
}

var _ driver.Rows = &revDiffRows{}

func (r *revDiffRows) Close() error {
	r.revmap = nil
	return nil
}

// maxRev returns the highest key from the map
func maxRev(revs map[string]struct{}) string {
	var max string
	for k := range revs {
		if k > max {
			max = k
		}
	}
	return max
}

func (r *revDiffRows) next() (docID string, missing []string, err error) {
	if len(r.revmap) == 0 {
		return "", nil, io.EOF
	}
	if err := r.ctx.Err(); err != nil {
		return "", nil, err
	}
	revs := map[string]struct{}{}
	for k, v := range r.revmap {
		docID = k
		for _, rev := range v {
			revs[rev] = struct{}{}
		}
		break
	}
	delete(r.revmap, docID)
	for len(revs) > 0 {
		rev := maxRev(revs)
		delete(revs, rev)
		doc, err := r.db.cdb.OpenDocID(docID, kivik.Rev(rev))
		if kivik.HTTPStatus(err) == http.StatusNotFound {
			missing = append(missing, rev)
			continue
		}
		if err != nil {
			return "", nil, err
		}
		for _, revid := range doc.Revisions[0].RevHistory.Ancestors()[1:] {
			delete(revs, revid)
		}
	}
	if len(missing) == 0 {
		return r.next()
	}
	return docID, missing, nil
}

func (r *revDiffRows) Next(row *driver.Row) error {
	docID, missing, err := r.next()
	if err != nil {
		return err
	}
	row.ID = docID
	value, err := json.Marshal(map[string][]string{
		"missing": missing,
	})
	row.Value = bytes.NewReader(value)
	return err
}

func (r *revDiffRows) Offset() int64     { return 0 }
func (r *revDiffRows) TotalRows() int64  { return 0 }
func (r *revDiffRows) UpdateSeq() string { return "" }
