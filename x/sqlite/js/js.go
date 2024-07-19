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

// Package js provides convenience wrappers around the goja JavaScript runtime.
package js

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/dop251/goja"
)

// MapFunc is the Go representation of a CouchDB [map function]. Exceptions are
// converted to errors.
//
// [map function]: https://docs.couchdb.org/en/stable/ddocs/views/nosql.html#map-functions
type MapFunc func(doc any) error

// Map compiles the provided JavaScript code into a MapFunc, and makes emit
// available to the JavaScript code.
func Map(code string, emit func(key, value any)) (MapFunc, error) {
	vm := goja.New()

	if err := vm.Set("emit", emit); err != nil {
		return nil, err
	}

	if _, err := vm.RunString("const map = " + code); err != nil {
		return nil, err
	}

	mapFunc, ok := goja.AssertFunction(vm.Get("map"))
	if !ok {
		panic(fmt.Sprintf("expected map to be a function, got %T", vm.Get("map")))
	}

	return func(doc any) error {
		if _, err := mapFunc(goja.Undefined(), vm.ToValue(doc)); err != nil {
			var exception *goja.Exception
			if errors.As(err, &exception) {
				return errors.New(exception.String())
			}
			// Should never happen
			return err
		}
		return nil
	}, nil
}

// FilterFunc represents a CouchDB [filter function]. Exceptions are converted
// to errors.
//
// [filter function]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#filter-functions
type FilterFunc func(doc, req any) (bool, error)

// Filter compiles the provided JavaScript code into a FilterFunc.
func Filter(code string) (FilterFunc, error) {
	vm := goja.New()
	if _, err := vm.RunString("const filter = " + code); err != nil {
		return nil, fmt.Errorf("failed to compile filter function: %s", err)
	}
	filterFunc, ok := goja.AssertFunction(vm.Get("filter"))
	if !ok {
		panic(fmt.Sprintf("expected filter to be a function, got %T", vm.Get("filter")))
	}
	return func(doc, req any) (bool, error) {
		result, err := filterFunc(goja.Undefined(), vm.ToValue(doc), vm.ToValue(req))
		if err != nil {
			var exception *goja.Exception
			if errors.As(err, &exception) {
				return false, errors.New(exception.String())
			}
			// Should never happen
			return false, err
		}
		rv := result.Export()
		b, _ := rv.(bool)
		return b, nil
	}, nil
}

// ReduceFunc is the Go representation of a CouchDB [reduce function]. Exceptions
// are converted to errors. The JavaScript function may return either a single
// item, or an array.  If a single item is returned, it is wrapped in an array
// before being returned to the caller.
//
// [reduce function]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#reduce-and-rereduce-functions
type ReduceFunc func(keys [][2]interface{}, values []interface{}, rereduce bool) ([]interface{}, error)

// Reduce compiles the provided JavaScript code into a ReduceFunc.
func Reduce(code string) (ReduceFunc, error) {
	vm := goja.New()

	if _, err := vm.RunString("const reduce = " + code); err != nil {
		return nil, err
	}
	reduceFunc, ok := goja.AssertFunction(vm.Get("reduce"))
	if !ok {
		return nil, fmt.Errorf("expected reduce to be a function, got %T", vm.Get("reduce"))
	}

	return func(keys [][2]interface{}, values []interface{}, rereduce bool) ([]interface{}, error) {
		reduceValue, err := reduceFunc(goja.Undefined(), vm.ToValue(keys), vm.ToValue(values), vm.ToValue(rereduce))
		if err != nil {
			var exception *goja.Exception
			if errors.As(err, &exception) {
				return nil, errors.New(exception.String())
			}
			// Should never happen
			return nil, err
		}

		rv := reduceValue.Export()
		// If rv is a slice, convert it to a []interface{} before returning it.
		v := reflect.ValueOf(rv)
		if v.Kind() == reflect.Slice {
			out := make([]interface{}, v.Len())
			for i := 0; i < v.Len(); i++ {
				out[i] = v.Index(i).Interface()
			}
			return out, nil
		}

		return []interface{}{rv}, nil
	}, nil
}
