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

package main

import (
	"context"
	"errors"
	"reflect"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

// method contains the relevant information for a driver method.
type method struct {
	// The method name
	Name string
	// Accepted values, except for context and options
	Accepts []reflect.Type
	// Return values, except for error
	Returns        []reflect.Type
	AcceptsContext bool
	AcceptsOptions bool
	ReturnsError   bool
	DBMethod       bool
}

var (
	typeContext       = reflect.TypeOf((*context.Context)(nil)).Elem()
	typeDriverOptions = reflect.TypeOf((*driver.Options)(nil)).Elem()
	typeClientOptions = reflect.TypeOf([]kivik.Option{})
	typeError         = reflect.TypeOf((*error)(nil)).Elem()
	typeString        = reflect.TypeOf("")
)

func parseMethods(input interface{}, isClient bool, skip map[string]struct{}) ([]*method, error) {
	var hasReceiver bool
	t := reflect.TypeOf(input)
	if t.Kind() != reflect.Struct {
		return nil, errors.New("input must be struct")
	}
	if t.NumField() != 1 || t.Field(0).Name != "X" {
		return nil, errors.New("wrapper struct must have a single field: X")
	}
	fType := t.Field(0).Type
	if isClient {
		if fType.Kind() != reflect.Ptr {
			return nil, errors.New("field X must be of type pointer to struct")
		}
		if fType.Elem().Kind() != reflect.Struct {
			return nil, errors.New("field X must be of type pointer to struct")
		}
		hasReceiver = true
	} else if fType.Kind() != reflect.Interface {
		return nil, errors.New("field X must be of type interface")
	}
	result := make([]*method, 0, fType.NumMethod())
	for i := 0; i < fType.NumMethod(); i++ {
		m := fType.Method(i)
		if _, ok := skip[m.Name]; ok {
			continue
		}
		dm := &method{
			Name: m.Name,
		}
		result = append(result, dm)
		accepts := make([]reflect.Type, m.Type.NumIn())
		for j := 0; j < m.Type.NumIn(); j++ {
			accepts[j] = m.Type.In(j)
		}
		if hasReceiver {
			accepts = accepts[1:]
		}
		if len(accepts) > 0 && accepts[0].Kind() == reflect.Interface && accepts[0].Implements(typeContext) {
			dm.AcceptsContext = true
			accepts = accepts[1:]
		}
		if !isClient && len(accepts) > 0 && accepts[len(accepts)-1] == typeDriverOptions {
			dm.AcceptsOptions = true
			accepts = accepts[:len(accepts)-1]
		}
		if isClient && m.Type.IsVariadic() && len(accepts) > 0 && accepts[len(accepts)-1].String() == typeClientOptions.String() {
			dm.AcceptsOptions = true
			accepts = accepts[:len(accepts)-1]
		}
		if len(accepts) > 0 {
			dm.Accepts = accepts
		}

		returns := make([]reflect.Type, m.Type.NumOut())
		for j := 0; j < m.Type.NumOut(); j++ {
			returns[j] = m.Type.Out(j)
		}
		if len(returns) > 0 && returns[len(returns)-1] == typeError {
			dm.ReturnsError = true
			returns = returns[:len(returns)-1]
		}
		if len(returns) > 0 {
			dm.Returns = returns
		}
	}
	return result, nil
}
