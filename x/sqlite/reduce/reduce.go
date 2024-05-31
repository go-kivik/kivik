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

// Package reduce implements CouchDB reduce function handling.
package reduce

// Row represents a single row of data to be reduced, or the result of a
// reduction. Key and Value are expected to represent JSON serializable data,
// and passing non-serializable data may result in a panic. Key and ID are only
// used for input rows as returned by a map function. Both are always empty for
// output rows.
type Row struct {
	ID    string
	Key   any
	Value any
}

// Func is the signature of a [CouchDB reduce function], translated to Go.
//
// [CouchDB reduce function]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#reduce-and-rereduce-functions
type Func func(keys [][2]interface{}, values []interface{}, rereduce bool) ([]interface{}, error)

// Reduce calls fn on rows, and returns the results.
func Reduce(rows []Row, fn Func) ([]Row, error) {
	if len(rows) == 0 {
		return nil, nil
	}
	keys := make([][2]interface{}, len(rows))
	values := make([]interface{}, len(rows))
	for i, row := range rows {
		keys[i] = [2]interface{}{row.Key, row.ID}
		values[i] = row.Value
	}
	results, err := fn(keys, values, false)
	if err != nil {
		return nil, err
	}
	out := make([]Row, len(results))
	for i, result := range results {
		out[i].Value = result
	}
	return out, nil
}
