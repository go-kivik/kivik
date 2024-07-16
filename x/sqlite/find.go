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

package sqlite

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/mango"
)

func parseQuery(query interface{}) (*mango.Selector, error) {
	var selector []byte
	switch t := query.(type) {
	case string:
		selector = []byte(t)
	case []byte:
		selector = t
	case json.RawMessage:
		selector = t
	default:
		var err error
		selector, err = json.Marshal(query)
		if err != nil {
			return nil, err
		}
	}
	var s mango.Selector
	err := json.Unmarshal(selector, &s)
	return &s, err
}

func (d *db) Find(ctx context.Context, query interface{}, options driver.Options) (driver.Rows, error) {
	_, err := parseQuery(query)
	if err != nil {
		return nil, &errors.Error{Status: http.StatusBadRequest, Err: err}
	}
	return d.Query(ctx, viewAllDocs, "", options)
}
