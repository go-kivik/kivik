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

package mock

import (
	"io"

	"github.com/go-kivik/kivik/v4/driver"
)

// Rows mocks driver.Rows
type Rows struct {
	// ID identifies a specific Rows instance.
	ID            string
	CloseFunc     func() error
	NextFunc      func(*driver.Row) error
	OffsetFunc    func() int64
	TotalRowsFunc func() int64
	UpdateSeqFunc func() string
}

var _ driver.Rows = &Rows{}

// Close calls r.CloseFunc
func (r *Rows) Close() error {
	if r.CloseFunc == nil {
		return nil
	}
	return r.CloseFunc()
}

// Next calls r.NextFunc
func (r *Rows) Next(row *driver.Row) error {
	if r.NextFunc == nil {
		return io.EOF
	}
	return r.NextFunc(row)
}

// Offset calls r.OffsetFunc
func (r *Rows) Offset() int64 {
	return r.OffsetFunc()
}

// TotalRows calls r.TotalRowsFunc
func (r *Rows) TotalRows() int64 {
	return r.TotalRowsFunc()
}

// UpdateSeq calls r.UpdateSeqFunc
func (r *Rows) UpdateSeq() string {
	return r.UpdateSeqFunc()
}

// RowsWarner wraps driver.RowsWarner
type RowsWarner struct {
	*Rows
	WarningFunc func() string
}

var _ driver.RowsWarner = &RowsWarner{}

// Warning calls r.WarningFunc
func (r *RowsWarner) Warning() string {
	return r.WarningFunc()
}

// Bookmarker wraps driver.Bookmarker
type Bookmarker struct {
	*Rows
	BookmarkFunc func() string
}

var _ driver.Bookmarker = &Bookmarker{}

// Bookmark calls r.BookmarkFunc
func (r *Bookmarker) Bookmark() string {
	return r.BookmarkFunc()
}

// QueryIndexer provides driver.QueryIndexer.
type QueryIndexer struct {
	*Rows
	QueryIndexFunc func() int
}

var _ driver.QueryIndexer = &QueryIndexer{}

// QueryIndex calls r.QueryIndexFunc
func (r *QueryIndexer) QueryIndex() int {
	return r.QueryIndexFunc()
}
