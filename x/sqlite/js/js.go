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
	"context"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"github.com/dop251/goja"
)

// Runtime holds configuration for JavaScript execution.
type Runtime struct {
	timeout time.Duration
}

// New creates a new Runtime with the given timeout. If timeout is zero, no
// timeout is enforced and cancellation depends entirely on the caller's context.
func New(timeout time.Duration) *Runtime {
	return &Runtime{timeout: timeout}
}

// MapFunc is the Go representation of a CouchDB [map function]. Exceptions are
// converted to errors. The context controls cancellation; if the context is
// cancelled, the VM is interrupted and the context error is returned.
//
// [map function]: https://docs.couchdb.org/en/stable/ddocs/views/nosql.html#map-functions
type MapFunc func(ctx context.Context, doc any) error

// Map compiles the provided JavaScript code into a MapFunc, and makes emit
// available to the JavaScript code. It uses a zero-value Runtime (no timeout).
func Map(code string, emit func(key, value any)) (MapFunc, error) {
	return new(Runtime).Map(code, emit)
}

// Map compiles the provided JavaScript code into a MapFunc, and makes emit
// available to the JavaScript code.
func (r *Runtime) Map(code string, emit func(key, value any)) (MapFunc, error) {
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

	return func(ctx context.Context, doc any) error {
		if r.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, r.timeout)
			defer cancel()
		}
		done := watchContext(ctx, vm)
		defer done()
		_, err := mapFunc(goja.Undefined(), vm.ToValue(doc))
		return exception(err)
	}, nil
}

// FilterFunc represents a CouchDB [filter function]. Exceptions are converted
// to errors. The context controls cancellation; if the context is cancelled,
// the VM is interrupted and the context error is returned.
//
// [filter function]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#filter-functions
type FilterFunc func(ctx context.Context, doc, req any) (bool, error)

// Filter compiles the provided JavaScript code into a FilterFunc.
// It uses a zero-value Runtime (no timeout).
func Filter(code string) (FilterFunc, error) {
	return new(Runtime).Filter(code)
}

// Filter compiles the provided JavaScript code into a FilterFunc.
func (r *Runtime) Filter(code string) (FilterFunc, error) {
	vm := goja.New()
	if _, err := vm.RunString("const filter = " + code); err != nil {
		return nil, fmt.Errorf("failed to compile filter function: %s", err)
	}
	filterFunc, ok := goja.AssertFunction(vm.Get("filter"))
	if !ok {
		panic(fmt.Sprintf("expected filter to be a function, got %T", vm.Get("filter")))
	}
	return func(ctx context.Context, doc, req any) (bool, error) {
		if r.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, r.timeout)
			defer cancel()
		}
		done := watchContext(ctx, vm)
		defer done()
		result, err := filterFunc(goja.Undefined(), vm.ToValue(doc), vm.ToValue(req))
		if err != nil {
			return false, exception(err)
		}
		rv := result.Export()
		b, _ := rv.(bool)
		return b, nil
	}, nil
}

// ReduceFunc is the Go representation of a CouchDB [reduce function]. Exceptions
// are converted to errors. The JavaScript function may return either a single
// item, or an array.  If a single item is returned, it is wrapped in an array
// before being returned to the caller. The context controls cancellation.
//
// [reduce function]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#reduce-and-rereduce-functions
type ReduceFunc func(ctx context.Context, keys [][2]any, values []any, rereduce bool) ([]any, error)

// Reduce compiles the provided JavaScript code into a ReduceFunc.
// It uses a zero-value Runtime (no timeout).
func Reduce(code string) (ReduceFunc, error) {
	return new(Runtime).Reduce(code)
}

// Reduce compiles the provided JavaScript code into a ReduceFunc.
func (r *Runtime) Reduce(code string) (ReduceFunc, error) {
	vm := goja.New()

	if _, err := vm.RunString("const reduce = " + code); err != nil {
		return nil, err
	}
	reduceFunc, ok := goja.AssertFunction(vm.Get("reduce"))
	if !ok {
		return nil, fmt.Errorf("expected reduce to be a function, got %T", vm.Get("reduce"))
	}

	return func(ctx context.Context, keys [][2]any, values []any, rereduce bool) ([]any, error) {
		if r.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, r.timeout)
			defer cancel()
		}
		done := watchContext(ctx, vm)
		defer done()
		reduceValue, err := reduceFunc(goja.Undefined(), vm.ToValue(keys), vm.ToValue(values), vm.ToValue(rereduce))
		if err != nil {
			return nil, exception(err)
		}

		rv := reduceValue.Export()
		// If rv is a slice, convert it to a []interface{} before returning it.
		v := reflect.ValueOf(rv)
		if v.Kind() == reflect.Slice {
			out := make([]any, v.Len())
			for i := 0; i < v.Len(); i++ {
				out[i] = v.Index(i).Interface()
			}
			return out, nil
		}

		return []any{rv}, nil
	}, nil
}

// ValidateFunc represents a CouchDB [validate_doc_update function]. The
// function throws to reject a document update. If the thrown value is an
// object with a "forbidden" key, the error will have HTTP status 403. If it
// has an "unauthorized" key, the error will have HTTP status 401.
//
// [validate_doc_update function]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#validate-document-update-functions
type ValidateFunc func(ctx context.Context, newDoc, oldDoc, userCtx, secObj any) error

// Validate compiles the provided JavaScript code into a ValidateFunc.
// It uses a zero-value Runtime (no timeout).
func Validate(code string) (ValidateFunc, error) {
	return new(Runtime).Validate(code)
}

// Validate compiles the provided JavaScript code into a ValidateFunc.
func (r *Runtime) Validate(code string) (ValidateFunc, error) {
	vm := goja.New()
	if _, err := vm.RunString("const validate = " + code); err != nil {
		return nil, fmt.Errorf("failed to compile validate function: %s", err)
	}
	validateFunc, ok := goja.AssertFunction(vm.Get("validate"))
	if !ok {
		panic(fmt.Sprintf("expected validate to be a function, got %T", vm.Get("validate")))
	}
	return func(ctx context.Context, newDoc, oldDoc, userCtx, secObj any) error {
		if r.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, r.timeout)
			defer cancel()
		}
		done := watchContext(ctx, vm)
		defer done()
		_, err := validateFunc(goja.Undefined(), vm.ToValue(newDoc), vm.ToValue(oldDoc), vm.ToValue(userCtx), vm.ToValue(secObj))
		if err != nil {
			return validateException(err)
		}
		return nil
	}, nil
}

// validateError represents an error returned by a validate_doc_update function.
type validateError struct {
	Status  int
	Message string
}

func (e *validateError) Error() string {
	return e.Message
}

// HTTPStatus returns the HTTP status code associated with the error.
func (e *validateError) HTTPStatus() int {
	return e.Status
}

func validateException(err error) error {
	var exc *goja.Exception
	if !errors.As(err, &exc) {
		return err
	}
	val := exc.Value().Export()
	if m, ok := val.(map[string]any); ok {
		if msg, ok := m["forbidden"]; ok {
			return &validateError{Status: http.StatusForbidden, Message: fmt.Sprint(msg)}
		}
		if msg, ok := m["unauthorized"]; ok {
			return &validateError{Status: http.StatusUnauthorized, Message: fmt.Sprint(msg)}
		}
		for _, v := range m {
			return &validateError{Status: http.StatusInternalServerError, Message: fmt.Sprint(v)}
		}
	}
	return &validateError{Status: http.StatusInternalServerError, Message: fmt.Sprint(val)}
}

// UpdateFunc represents a CouchDB [update function]. It accepts a document and
// a request object and returns the updated document and a response string.
// The context controls cancellation; if the context is cancelled, the VM is
// interrupted and the context error is returned.
//
// [update function]: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#update-functions
type UpdateFunc func(ctx context.Context, doc, req any) (any, string, error)

// Update compiles the provided JavaScript code into an UpdateFunc.
// It uses a zero-value Runtime (no timeout).
func Update(code string) (UpdateFunc, error) {
	return new(Runtime).Update(code)
}

// Update compiles the provided JavaScript code into an UpdateFunc.
func (r *Runtime) Update(code string) (UpdateFunc, error) {
	vm := goja.New()
	if _, err := vm.RunString("const update = " + code); err != nil {
		return nil, fmt.Errorf("failed to compile update function: %s", err)
	}
	updateFunc, ok := goja.AssertFunction(vm.Get("update"))
	if !ok {
		panic(fmt.Sprintf("expected update to be a function, got %T", vm.Get("update")))
	}
	return func(ctx context.Context, doc, req any) (any, string, error) {
		if r.timeout > 0 {
			var cancel context.CancelFunc
			ctx, cancel = context.WithTimeout(ctx, r.timeout)
			defer cancel()
		}
		done := watchContext(ctx, vm)
		defer done()
		result, err := updateFunc(goja.Undefined(), vm.ToValue(doc), vm.ToValue(req))
		if err != nil {
			return nil, "", exception(err)
		}
		rv := result.Export()
		arr, _ := rv.([]any)
		if len(arr) != 2 {
			return nil, "", fmt.Errorf("update function must return [doc, response], got %v", rv)
		}
		resp, _ := arr[1].(string)
		return arr[0], resp, nil
	}, nil
}

// watchContext starts a goroutine that interrupts the VM when the context is
// done. The returned function must be called (via defer) to clean up. It also
// calls ClearInterrupt to ensure the VM is reusable.
func watchContext(ctx context.Context, vm *goja.Runtime) func() {
	ch := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			vm.Interrupt(ctx.Err())
		case <-ch:
		}
	}()
	return func() {
		close(ch)
		vm.ClearInterrupt()
	}
}

// exception converts a JavaScript exception to a Go error. If the error is
// a [goja.InterruptedError], the underlying cause (e.g. context error) is
// unwrapped and returned directly.
func exception(err error) error {
	if err == nil {
		return nil
	}
	var interrupted *goja.InterruptedError
	if errors.As(err, &interrupted) {
		if cause, ok := interrupted.Value().(error); ok {
			return cause
		}
		return interrupted
	}
	var exc *goja.Exception
	if errors.As(err, &exc) {
		return errors.New(exc.String())
	}
	// goja should only return *Exception or *InterruptedError
	return err
}
