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

package couchdb

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
)

const (
	pathIndex = "_index"
)

func (d *db) CreateIndex(ctx context.Context, ddoc, name string, index interface{}, options driver.Options) error {
	opts := map[string]interface{}{}
	options.Apply(opts)
	reqPath := partPath(pathIndex)
	options.Apply(reqPath)
	indexObj, err := deJSONify(index)
	if err != nil {
		return err
	}
	parameters := struct {
		Index interface{} `json:"index"`
		Ddoc  string      `json:"ddoc,omitempty"`
		Name  string      `json:"name,omitempty"`
	}{
		Index: indexObj,
		Ddoc:  ddoc,
		Name:  name,
	}
	chttpOpts := &chttp.Options{
		Body: chttp.EncodeBody(parameters),
	}
	_, err = d.Client.DoError(ctx, http.MethodPost, d.path(reqPath.String()), chttpOpts)
	return err
}

func (d *db) GetIndexes(ctx context.Context, options driver.Options) ([]driver.Index, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	reqPath := partPath(pathIndex)
	options.Apply(reqPath)
	var result struct {
		Indexes []driver.Index `json:"indexes"`
	}
	err := d.Client.DoJSON(ctx, http.MethodGet, d.path(reqPath.String()), nil, &result)
	return result.Indexes, err
}

func (d *db) DeleteIndex(ctx context.Context, ddoc, name string, options driver.Options) error {
	opts := map[string]interface{}{}
	options.Apply(opts)
	if ddoc == "" {
		return missingArg("ddoc")
	}
	if name == "" {
		return missingArg("name")
	}
	reqPath := partPath(pathIndex)
	options.Apply(reqPath)
	path := fmt.Sprintf("%s/%s/json/%s", reqPath, ddoc, name)
	_, err := d.Client.DoError(ctx, http.MethodDelete, d.path(path), nil)
	return err
}

func (d *db) Find(ctx context.Context, query interface{}, options driver.Options) (driver.Rows, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	reqPath := partPath("_find")
	options.Apply(reqPath)
	chttpOpts := &chttp.Options{
		GetBody: chttp.BodyEncoder(query),
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	resp, err := d.Client.DoReq(ctx, http.MethodPost, d.path(reqPath.String()), chttpOpts)
	if err != nil {
		return nil, err
	}
	if err = chttp.ResponseError(resp); err != nil {
		return nil, err
	}
	return newFindRows(ctx, resp.Body), nil
}

type queryPlan struct {
	DBName   string                 `json:"dbname"`
	Index    map[string]interface{} `json:"index"`
	Selector map[string]interface{} `json:"selector"`
	Options  map[string]interface{} `json:"opts"`
	Limit    int64                  `json:"limit"`
	Skip     int64                  `json:"skip"`
	Fields   fields                 `json:"fields"`
	Range    map[string]interface{} `json:"range"`
}

type fields []interface{}

func (f *fields) UnmarshalJSON(data []byte) error {
	if string(data) == `"all_fields"` {
		return nil
	}
	var i []interface{}
	if err := json.Unmarshal(data, &i); err != nil {
		return err
	}
	newFields := make([]interface{}, len(i))
	copy(newFields, i)
	*f = newFields
	return nil
}

func (d *db) Explain(ctx context.Context, query interface{}, options driver.Options) (*driver.QueryPlan, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	reqPath := partPath("_explain")
	options.Apply(reqPath)
	chttpOpts := &chttp.Options{
		GetBody: chttp.BodyEncoder(query),
		Header: http.Header{
			chttp.HeaderIdempotencyKey: []string{},
		},
	}
	var plan queryPlan
	if err := d.Client.DoJSON(ctx, http.MethodPost, d.path(reqPath.String()), chttpOpts, &plan); err != nil {
		return nil, err
	}
	return &driver.QueryPlan{
		DBName:   plan.DBName,
		Index:    plan.Index,
		Selector: plan.Selector,
		Options:  plan.Options,
		Limit:    plan.Limit,
		Skip:     plan.Skip,
		Fields:   plan.Fields,
		Range:    plan.Range,
	}, nil
}
