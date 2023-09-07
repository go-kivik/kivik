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
	"context"
	"fmt"

	"github.com/gopherjs/gopherjs/js"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

type bulkResult struct {
	*js.Object
	OK         bool   `js:"ok"`
	ID         string `js:"id"`
	Rev        string `js:"rev"`
	Error      string `js:"name"`
	StatusCode int    `js:"status"`
	Reason     string `js:"message"`
	IsError    bool   `js:"error"`
}

func (d *db) BulkDocs(ctx context.Context, docs []interface{}, options driver.Options) (results []driver.BulkResult, err error) {
	defer func() {
		if r := recover(); r != nil {
			if e, ok := r.(error); ok {
				err = e
			} else {
				err = fmt.Errorf("%v", r)
			}
		}
	}()
	opts := map[string]interface{}{}
	options.Apply(opts)
	result, err := d.db.BulkDocs(ctx, docs, opts)
	if err != nil {
		return nil, err
	}
	for result != js.Undefined && result.Length() > 0 {
		r := &bulkResult{}
		r.Object = result.Call("shift")
		var err error
		if r.IsError {
			err = &kivik.Error{Status: r.StatusCode, Message: r.Reason}
		}
		results = append(results, driver.BulkResult{
			ID:    r.ID,
			Rev:   r.Rev,
			Error: err,
		})
	}

	return results, nil
}
