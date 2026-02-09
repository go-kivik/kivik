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

package registry

import (
	"sync"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

// to protect the registry from concurrent tests
var registryMU sync.Mutex

func TestRegister(t *testing.T) {
	registryMU.Lock()
	defer registryMU.Unlock()
	t.Run("nil driver", func(t *testing.T) {
		t.Cleanup(func() {
			drivers = make(map[string]driver.Driver)
		})
		p := func() (p any) {
			defer func() {
				p = recover()
			}()
			Register("foo", nil)
			return ""
		}()
		if p.(string) != "kivik: Register driver is nil" {
			t.Errorf("Unexpected panic: %v", p)
		}
	})

	t.Run("duplicate driver", func(t *testing.T) {
		t.Cleanup(func() {
			drivers = make(map[string]driver.Driver)
		})
		p := func() (p any) {
			defer func() {
				p = recover()
			}()
			Register("foo", &mock.Driver{})
			Register("foo", &mock.Driver{})
			return ""
		}()
		if p.(string) != "kivik: Register called twice for driver foo" {
			t.Errorf("Unexpected panic: %v", p)
		}
	})

	t.Run("success", func(t *testing.T) {
		t.Cleanup(func() {
			drivers = make(map[string]driver.Driver)
		})
		p := func() (p any) {
			defer func() {
				p = recover()
			}()
			Register("foo", &mock.Driver{})
			return ""
		}()
		if p != nil {
			t.Errorf("Unexpected panic: %v", p)
		}
		expected := map[string]driver.Driver{
			"foo": &mock.Driver{},
		}
		if d := testy.DiffInterface(expected, drivers); d != nil {
			t.Error(d)
		}
	})
}
