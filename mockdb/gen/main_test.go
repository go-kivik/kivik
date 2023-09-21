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
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestCompareMethods(t *testing.T) {
	type tst struct {
		client    []*method
		driver    []*method
		expSame   []*method
		expClient []*method
		expDriver []*method
	}
	tests := testy.NewTable()
	tests.Add("one identical", tst{
		client: []*method{
			{Name: "Foo"},
		},
		driver: []*method{
			{Name: "Foo"},
		},
		expSame: []*method{
			{Name: "Foo"},
		},
		expClient: []*method{},
		expDriver: []*method{},
	})
	tests.Add("same name", tst{
		client: []*method{
			{Name: "Foo", ReturnsError: true},
		},
		driver: []*method{
			{Name: "Foo"},
		},
		expSame: []*method{},
		expClient: []*method{
			{Name: "Foo", ReturnsError: true},
		},
		expDriver: []*method{
			{Name: "Foo"},
		},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		same, client, driver := compareMethods(test.client, test.driver)
		if d := testy.DiffInterface(test.expSame, same); d != nil {
			t.Errorf("Same:\n%s\n", d)
		}
		if d := testy.DiffInterface(test.expClient, client); d != nil {
			t.Errorf("Same:\n%s\n", d)
		}
		if d := testy.DiffInterface(test.expDriver, driver); d != nil {
			t.Errorf("Same:\n%s\n", d)
		}
	})
}
