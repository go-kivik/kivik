package kivik

import (
	"testing"

	"github.com/flimzy/kivik/driver"
)

type fakeRep struct {
	driver.Replication
	state string
}

func (r *fakeRep) State() string {
	return r.state
}

func TestReplicationIsActive(t *testing.T) {
	t.Run("Active", func(t *testing.T) {
		r := &Replication{
			irep: &fakeRep{state: "active"},
		}
		if !r.IsActive() {
			t.Errorf("Expected active")
		}
	})
	t.Run("Complete", func(t *testing.T) {
		r := &Replication{
			irep: &fakeRep{state: string(ReplicationComplete)},
		}
		if r.IsActive() {
			t.Errorf("Expected not active")
		}
	})
	t.Run("Nil", func(t *testing.T) {
		var r *Replication
		if r.IsActive() {
			t.Errorf("Expected not active")
		}
	})
}

func TestReplicationDocsWritten(t *testing.T) {
	t.Run("No Info", func(t *testing.T) {
		r := &Replication{}
		result := r.DocsWritten()
		if result != 0 {
			t.Errorf("Unexpected doc count: %d", result)
		}
	})
	t.Run("With Info", func(t *testing.T) {
		r := &Replication{
			info: &driver.ReplicationInfo{
				DocsWritten: 123,
			},
		}
		result := r.DocsWritten()
		if result != 123 {
			t.Errorf("Unexpected doc count: %d", result)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		var r *Replication
		result := r.DocsWritten()
		if result != 0 {
			t.Errorf("Unexpected doc count: %d", result)
		}
	})
}
