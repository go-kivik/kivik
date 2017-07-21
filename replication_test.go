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
