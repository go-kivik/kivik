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

import (
	"io"
	"log"
	"reflect"
)

// RowIterator is the interface for iterating over rows of data to be reduced.
type RowIterator interface {
	// Next should populate Row, or return an error. It should return [io.EOF]
	// when there are no more rows to read.
	Next(*Row) error
}

// Row represents a single row of data to be reduced, or the result of a
// reduction. Key and Value are expected to represent JSON serializable data,
// and passing non-serializable data may result in a panic. ID is only used for
// input rows as returned by a map function. It is always empty for output rows.
type Row struct {
	// First and Last reference the key's primary key, and are used to
	// disambiguate rows with the same key. For map inputs, they should be
	// the same.  For reduced inputs, they represent a range of keys.
	First int
	Last  int
	ID    string
	Key   any
	Value any
}

// Rows is a slice of Row, and implements RowIterator.
type Rows []Row

var _ RowIterator = (*Rows)(nil)

// Next implements RowIterator.
func (r *Rows) Next(row *Row) error {
	if len(*r) == 0 {
		return io.EOF
	}
	*row, *r = (*r)[0], (*r)[1:]
	return nil
}

// Func is the signature of a [CouchDB reduce function], translated to Go.
//
// [CouchDB reduce function]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#reduce-and-rereduce-functions
type Func func(keys [][2]interface{}, values []interface{}, rereduce bool) ([]interface{}, error)

// Reduce calls fn on rows, and returns the results. The input must be in
// key-sorted order, and may contain both previously reduced rows, and map
// output rows.  cb, if not nil, is called with the results of every
// intermediate reduce step.
//
// The Key field of the returned row(s) will be set only when grouping.
//
// groupLevel controls grouping.  Possible values:
//
//	-1: Maximum grouping, same as group=true
//	 0: No grouping, same as group=false
//	1+: Group by the first N elements of the key, same as group_level=N
func Reduce(rows RowIterator, javascript string, logger *log.Logger, groupLevel int, cb func([]Row)) ([]Row, error) {
	fn, err := ParseFunc(javascript, logger)
	if err != nil {
		return nil, err
	}
	return reduce(rows, fn, groupLevel, cb)
}

func reduce(rows RowIterator, fn Func, groupLevel int, cb func([]Row)) ([]Row, error) {
	out := make([]Row, 0, 1)
	var first, last int

	callReduce := func(keys [][2]interface{}, values []interface{}, rereduce bool, key any) error {
		if len(keys) == 0 {
			return nil
		}
		if len(keys) == 1 && rereduce {
			// Nothing to rereduce if we have only a single input--just pass it through
			out = append(out, Row{
				Key:   key,
				Value: values[0],
				First: first,
				Last:  last,
			})
			return nil
		}
		results, err := fn(keys, values, rereduce)
		if err != nil {
			return err
		}
		rows := make([]Row, 0, len(results))
		for _, result := range results {
			row := Row{
				Value: result,
				First: first,
				Last:  last,
			}
			if keyLen(key) > 0 {
				row.Key = key
			}
			rows = append(rows, row)
			first, last = 0, 0
		}
		if cb != nil {
			cb(rows)
		}
		out = append(out, rows...)
		return nil
	}

	const defaultCap = 10
	keys := make([][2]interface{}, 0, defaultCap)
	values := make([]interface{}, 0, defaultCap)
	var targetKey any
	var rereduce bool
	for {
		var row Row
		if err := rows.Next(&row); err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if groupLevel != 0 {
			switch {
			case targetKey != nil && (!reflect.DeepEqual(targetKey, truncateKey(row.Key, groupLevel)) || rereduce != (row.ID == "")):
				if err := callReduce(keys, values, rereduce, targetKey); err != nil {
					return nil, err
				}

				keys = keys[:0]
				values = values[:0]
				fallthrough
			case targetKey == nil:
				targetKey = truncateKey(row.Key, groupLevel)
				rereduce = row.ID == ""
			}
		}
		if first == 0 {
			first = row.First
		}
		last = row.Last

		keys = append(keys, [2]interface{}{row.Key, row.ID})
		values = append(values, row.Value)
	}

	if err := callReduce(keys, values, rereduce, targetKey); err != nil {
		return nil, err
	}

	if len(out) <= 1 {
		// One or fewer results can't have duplicates that need to be re-reduced.
		return out, nil
	}

	// If we received mixed map/reduce inputs, then we may need to re-reduce
	// the output before returning.
	lastKey := truncateKey(out[0].Key, groupLevel)
	for i := 1; i < len(out); i++ {
		key := truncateKey(out[i].Key, groupLevel)
		if reflect.DeepEqual(lastKey, key) {
			rowsOut := Rows(out)
			return reduce(&rowsOut, fn, groupLevel, cb)
		}
	}

	return out, nil
}

func keyLen(key any) int {
	if key == nil {
		return 0
	}
	if k, ok := key.([]any); ok {
		return len(k)
	}
	return 1
}

// truncateKey truncates the key to the given level.
func truncateKey(key any, level int) any {
	if level == 0 {
		return nil
	}
	target, ok := key.([]any)
	if !ok {
		return key
	}

	if level > 0 && level < len(target) {
		return target[:level]
	}
	return target
}
