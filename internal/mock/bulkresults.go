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

import "github.com/go-kivik/kivik/v4/driver"

// BulkResults mocks driver.BulkResults
type BulkResults struct {
	// ID identifies a specific BulkResults instance
	ID        string
	NextFunc  func(*driver.BulkResult) error
	CloseFunc func() error
}

var _ driver.BulkResults = &BulkResults{}

// Next calls r.NextFunc
func (r *BulkResults) Next(result *driver.BulkResult) error {
	return r.NextFunc(result)
}

// Close calls r.CloseFunc
func (r *BulkResults) Close() error {
	return r.CloseFunc()
}
