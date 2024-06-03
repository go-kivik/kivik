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
	"errors"
	"net/http"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func (d *db) BulkDocs(ctx context.Context, docs []interface{}, options driver.Options) ([]driver.BulkResult, error) {
	chttpOpts := chttp.NewOptions(options)
	opts := map[string]interface{}{}
	options.Apply(opts)
	opts["docs"] = docs
	chttpOpts.GetBody = chttp.BodyEncoder(opts)

	resp, err := d.Client.DoReq(ctx, http.MethodPost, d.path("/_bulk_docs"), chttpOpts)
	if err != nil {
		return nil, err
	}
	defer chttp.CloseBody(resp.Body)

	switch resp.StatusCode {
	case http.StatusCreated:
		// Nothing to do
	case http.StatusExpectationFailed:
		err = &chttp.HTTPError{
			Response: resp,
			Reason:   "one or more document was rejected",
		}
	default:
		// All other errors can consume the response body and return immediately
		if e := chttp.ResponseError(resp); e != nil {
			return nil, e
		}
	}
	var temp []bulkDocResult
	if err := chttp.DecodeJSON(resp, &temp); err != nil {
		return nil, err
	}
	results := make([]driver.BulkResult, len(temp))
	for i, r := range temp {
		results[i] = driver.BulkResult(r)
	}
	return results, err
}

type bulkDocResult struct {
	ID    string `json:"id"`
	Rev   string `json:"rev"`
	Error error
}

func (r *bulkDocResult) UnmarshalJSON(p []byte) error {
	target := struct {
		*bulkDocResult
		Error         string `json:"error"`
		Reason        string `json:"reason"`
		UnmarshalJSON struct{}
	}{
		bulkDocResult: r,
	}
	if err := json.Unmarshal(p, &target); err != nil {
		return err
	}
	switch target.Error {
	case "":
		// No error
	case "conflict":
		r.Error = &internal.Error{Status: http.StatusConflict, Err: errors.New(target.Reason)}
	default:
		r.Error = &internal.Error{Status: http.StatusInternalServerError, Err: errors.New(target.Reason)}
	}
	return nil
}
