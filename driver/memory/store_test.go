package memory

import (
	"strings"
	"testing"

	"github.com/flimzy/diff"
)

func TestRandStr(t *testing.T) {
	str := randStr()
	if len(str) != 32 {
		t.Errorf("Expected 32-char string, got %d", len(str))
	}
}

func TestAddRevision(t *testing.T) {
	d := &database{
		docs: make(map[string]*document),
	}
	r := d.addRevision(jsondoc{"_id": "bar"})
	if !strings.HasPrefix(r, "1-") {
		t.Errorf("Expected initial revision to start with '1-', but got '%s'", r)
	}
	if len(r) != 34 {
		t.Errorf("rev (%s) is %d chars long, expected 34", r, len(r))
	}
	r = d.addRevision(jsondoc{"_id": "bar"})
	if !strings.HasPrefix(r, "2-") {
		t.Errorf("Expected second revision to start with '2-', but got '%s'", r)
	}
	if len(r) != 34 {
		t.Errorf("rev (%s) is %d chars long, expected 34", r, len(r))
	}
}

func TestAddLocalRevision(t *testing.T) {
	d := &database{
		docs: make(map[string]*document),
	}
	r := d.addRevision(jsondoc{"_id": "_local/foo"})
	if r != "1-0" {
		t.Errorf("Expected local revision, got %s", r)
	}
	r = d.addRevision(jsondoc{"_id": "_local/foo"})
	if r != "1-0" {
		t.Errorf("Expected local revision, got %s", r)
	}
}

func TestGetRevisionMissing(t *testing.T) {
	d := &database{
		docs: make(map[string]*document),
	}
	_, found := d.getRevision("foo", "bar")
	if found {
		t.Errorf("Should not have found missing revision")
	}
}

func TestGetRevisionFound(t *testing.T) {
	d := &database{
		docs: make(map[string]*document),
	}
	r := d.addRevision(map[string]interface{}{"_id": "foo", "a": 1})
	_ = d.addRevision(map[string]interface{}{"_id": "foo", "a": 2})
	result, found := d.getRevision("foo", r)
	if !found {
		t.Errorf("Should have found revision")
	}
	expected := map[string]interface{}{"_id": "foo", "a": 1, "_rev": r}
	if d := diff.AsJSON(expected, result.data); d != "" {
		t.Error(d)
	}
}

func TestRev(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		d := jsondoc{}
		if d.Rev() != "" {
			t.Errorf("Rev should be missing, but got %s", d.Rev())
		}
	})
	t.Run("Set", func(t *testing.T) {
		d := jsondoc{"_rev": "foo"}
		if d.Rev() != "foo" {
			t.Errorf("Rev should be foo, but got %s", d.Rev())
		}
	})
	t.Run("NonString", func(t *testing.T) {
		d := jsondoc{"_rev": true}
		if d.Rev() != "" {
			t.Errorf("Rev should be missing, but got %s", d.Rev())
		}
	})
}

func TestID(t *testing.T) {
	t.Run("Missing", func(t *testing.T) {
		d := jsondoc{}
		if d.ID() != "" {
			t.Errorf("ID should be missing, but got %s", d.ID())
		}
	})
	t.Run("Set", func(t *testing.T) {
		d := jsondoc{"_id": "foo"}
		if d.ID() != "foo" {
			t.Errorf("ID should be foo, but got %s", d.ID())
		}
	})
	t.Run("NonString", func(t *testing.T) {
		d := jsondoc{"_id": true}
		if d.ID() != "" {
			t.Errorf("ID should be missing, but got %s", d.ID())
		}
	})
}
