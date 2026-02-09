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
	"reflect"
	"testing"

	"gitlab.com/flimzy/testy"
)

func init() {
	initTemplates("templates")
}

func TestRenderExpectedType(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("CreateDoc", &method{
		Name:           "CreateDoc",
		DBMethod:       true,
		AcceptsContext: true,
		AcceptsOptions: true,
		ReturnsError:   true,
		Accepts:        []reflect.Type{reflect.TypeOf((*any)(nil)).Elem()},
		Returns:        []reflect.Type{typeString, typeString},
	})

	tests.Run(t, func(t *testing.T, m *method) {
		result, err := renderExpectedType(m)
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffText(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

func TestRenderDriverMethod(t *testing.T) {
	type tst struct {
		method *method
		err    string
	}
	tests := testy.NewTable()
	tests.Add("CreateDB", tst{
		method: &method{
			Name:           "CreateDB",
			Accepts:        []reflect.Type{typeString},
			AcceptsContext: true,
			AcceptsOptions: true,
			ReturnsError:   true,
		},
	})
	tests.Add("No context", tst{
		method: &method{
			Name:           "NoCtx",
			AcceptsOptions: true,
			ReturnsError:   true,
		},
	})
	tests.Run(t, func(t *testing.T, test tst) {
		result, err := renderDriverMethod(test.method)
		if !testy.ErrorMatches(test.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if d := testy.DiffText(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

func TestVariables(t *testing.T) {
	type tst struct {
		method   *method
		indent   int
		expected string
	}
	tests := testy.NewTable()
	tests.Add("no args", tst{
		method:   &method{},
		expected: "",
	})
	tests.Add("one arg", tst{
		method:   &method{Accepts: []reflect.Type{typeString}},
		expected: "arg0: arg0,",
	})
	tests.Add("one arg + options", tst{
		method: &method{Accepts: []reflect.Type{typeString}, AcceptsOptions: true},
		expected: `arg0:    arg0,
options: options,`,
	})
	tests.Add("indent", tst{
		method: &method{Accepts: []reflect.Type{typeString, typeString}, AcceptsOptions: true},
		indent: 2,
		expected: `		arg0:    arg0,
		arg1:    arg1,
		options: options,`,
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result := test.method.Variables(test.indent)
		if d := testy.DiffText(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}
