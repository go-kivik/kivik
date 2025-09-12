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

package options

import (
	"fmt"
	"net/http"

	internal "github.com/go-kivik/kivik/v4/int/errors"
)

// PaginationOptions are all of the options recognized by [_all_dbs], and
// views.
//
// [_all_dbs]: https://docs.couchdb.org/en/stable/api/server/common.html#all-dbs
type PaginationOptions struct {
	limit        int64
	skip         int64
	descending   bool
	endkey       string
	startkey     string
	inclusiveEnd bool
}

// PaginationOptions returns the pagination options for _all_dbs or a view.
func (o Map) PaginationOptions() (*PaginationOptions, error) {
	limit, err := o.Limit()
	if err != nil {
		return nil, err
	}
	skip, err := o.Skip()
	if err != nil {
		return nil, err
	}
	descending, err := o.Descending()
	if err != nil {
		return nil, err
	}
	inclusiveEnd, err := o.InclusiveEnd()
	if err != nil {
		return nil, err
	}
	endkey, err := o.EndKey()
	if err != nil {
		return nil, err
	}
	startkey, err := o.StartKey()
	if err != nil {
		return nil, err
	}

	return &PaginationOptions{
		limit:        limit,
		skip:         skip,
		descending:   descending,
		endkey:       endkey,
		startkey:     startkey,
		inclusiveEnd: inclusiveEnd,
	}, nil
}

func (p PaginationOptions) descendingModifier() int {
	if p.descending {
		return -1
	}
	return 1
}

// Validate returns an error if the options are invalid.
func (p PaginationOptions) Validate() error {
	if p.endkey != "" && p.startkey != "" && couchdbCmpString(p.startkey, p.endkey)*p.descendingModifier() > 0 {
		return &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("no rows can match your key range, reverse your start_key and end_key or set descending=%v", !p.descending)}
	}
	return nil
}

// BuildWhere returns WHERE conditions based on the provided configuration
// arguments, and may append to args as needed.
func (p PaginationOptions) BuildWhere(args *[]any) []string {
	where := make([]string, 0, defaultWhereCap)
	if p.endkey != "" {
		op := endKeyOp(p.descending, p.inclusiveEnd)
		where = append(where, fmt.Sprintf("view.key %s $%d", op, len(*args)+1))
		*args = append(*args, p.endkey)
	}
	if p.startkey != "" {
		op := startKeyOp(p.descending)
		where = append(where, fmt.Sprintf("view.key %s $%d", op, len(*args)+1))
		*args = append(*args, p.startkey)
	}
	return where
}

// BuildOrderBy returns an ORDER BY clause based on the provided configuration.
func (p PaginationOptions) BuildOrderBy() string {
	if p.descending {
		return "ORDER BY view.key DESC"
	}
	return "ORDER BY view.key ASC"
}
