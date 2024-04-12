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
	"io"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func toRevDiffRequest(revMap interface{}) (map[string][]string, error) {
	data, err := json.Marshal(revMap)
	if err != nil {
		return nil, &internal.Error{Message: "invalid body", Status: http.StatusBadRequest}
	}
	var req map[string][]string
	if err := json.Unmarshal(data, &req); err != nil {
		return nil, &internal.Error{Message: "invalid body", Status: http.StatusBadRequest}
	}
	return req, nil
}

func (db) RevsDiff(_ context.Context, revMap interface{}) (driver.Rows, error) {
	_, err := toRevDiffRequest(revMap)
	if err != nil {
		return nil, err
	}
	return &revsDiffResponse{}, nil
}

type revsDiffResponse struct{}

var _ driver.Rows = (*revsDiffResponse)(nil)

func (r *revsDiffResponse) Next(*driver.Row) error {
	return io.EOF
}

func (r *revsDiffResponse) Close() error {
	return nil
}

func (*revsDiffResponse) Offset() int64     { return 0 }
func (*revsDiffResponse) TotalRows() int64  { return 0 }
func (*revsDiffResponse) UpdateSeq() string { return "" }
