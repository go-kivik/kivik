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
