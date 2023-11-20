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

//go:build !js
// +build !js

package couchserver

import (
	"testing"

	"github.com/go-kivik/kivik/v4"
)

func TestVendor(t *testing.T) {
	t.Run("Unset", func(t *testing.T) {
		h := &Handler{}
		c, v, vv := h.vendor()
		if c != CompatVersion {
			t.Errorf("CompatVer Expected: %s\n  CompatVer Actual: %s", CompatVersion, c)
		}
		if want := "Kivik"; v != want {
			t.Errorf("Vendor Expected: %s\n  Vendor Actual: %s", want, v)
		}
		if vv != kivik.Version {
			t.Errorf("Vendor Version Expected: %s\n  Vendor Version Actual: %s", kivik.Version, vv)
		}
	})
	t.Run("Set", func(t *testing.T) {
		ec, ev, evv := "123.Foo", "Test", "123.Bar"
		h := &Handler{
			CompatVersion: ec,
			Vendor:        ev,
			VendorVersion: evv,
		}
		c, v, vv := h.vendor()
		if c != ec {
			t.Errorf("CompatVer Expected: %s\n  CompatVer Actual: %s", ec, c)
		}
		if v != ev {
			t.Errorf("Vendor Expected: %s\n  Vendor Actual: %s", ev, v)
		}
		if vv != evv {
			t.Errorf("Vendor Version Expected: %s\n  Vendor Version Actual: %s", evv, vv)
		}
	})
}
