package mockdb

import (
	"errors"
	"testing"
	"time"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/google/go-cmp/cmp"
)

func equateErrorMessages() cmp.Option {
	return cmp.Comparer(func(x, y error) bool {
		if x == nil || y == nil {
			return x == nil && y == nil
		}
		return x.Error() == y.Error()
	})
}

func TestReplicationMetadta(t *testing.T) {
	_, m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ts := time.Now()
	const (
		eID     = "a1"
		eSource = "a2"
		eTarget = "a3"
		eState  = kivik.ReplicationComplete
		eErr    = "a5"
	)
	eStartTime := ts
	eEndTime := ts.Add(time.Second)
	r := m.NewReplication().
		ID(eID).
		Source(eSource).
		Target(eTarget).
		StartTime(eStartTime).
		EndTime(eEndTime).
		State(eState).
		Err(errors.New(eErr))
	dr := &driverReplication{r}

	want := driver.ReplicationMetadata{
		ID:        eID,
		Source:    eSource,
		Target:    eTarget,
		StartTime: eStartTime,
		EndTime:   eEndTime,
		State:     string(eState),
		Error:     errors.New(eErr),
	}
	got := dr.Metadata()
	if d := cmp.Diff(want, got, equateErrorMessages()); d != "" {
		t.Errorf("unexpected metadata: %s", d)
	}
}
