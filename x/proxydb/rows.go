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

package proxydb

import (
	"bytes"
	"encoding/json"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

type rows struct {
	*kivik.ResultSet
}

var _ driver.Rows = &rows{}

func (r *rows) Next(row *driver.Row) error {
	if !r.ResultSet.Next() {
		return r.Err()
	}
	var value json.RawMessage
	if err := r.ScanValue(&value); err != nil {
		return err
	}
	var doc json.RawMessage
	if err := r.ScanDoc(&doc); err != nil {
		return err
	}
	var err error
	row.ID, err = r.ID()
	if err != nil {
		return err
	}
	key, err := r.Key()
	if err != nil {
		return err
	}
	row.Key = json.RawMessage(key)
	row.Value = bytes.NewReader(value)
	row.Doc = bytes.NewReader(doc)
	return nil
}

func (r *rows) Close() error {
	return r.ResultSet.Close()
}

func (r *rows) Offset() int64 {
	md, err := r.Metadata()
	if err != nil {
		return 0
	}
	return md.Offset
}

func (r *rows) TotalRows() int64 {
	md, err := r.Metadata()
	if err != nil {
		return 0
	}
	return md.TotalRows
}

func (r *rows) UpdateSeq() string {
	md, err := r.Metadata()
	if err != nil {
		return ""
	}
	return md.UpdateSeq
}
