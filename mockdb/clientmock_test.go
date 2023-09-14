package mockdb

import (
	"errors"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func TestReplication(t *testing.T) {
	_, m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ts := time.Now()
	const (
		eID     = "a1"
		eSource = "a2"
		eTarget = "a3"
		eErr    = "a5"
	)
	eStartTime := ts
	eEndTime := ts.Add(time.Second)
	eState := kivik.ReplicationComplete
	r := m.NewReplication().
		Metadata(driver.ReplicationMetadata{
			ID:        eID,
			Source:    eSource,
			Target:    eTarget,
			StartTime: eStartTime,
			EndTime:   eEndTime,
			State:     string(eState),
		}).
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
	}
	got := dr.Metadata()
	if d := cmp.Diff(want, got); d != "" {
		t.Error(d)
	}
	t.Run("Err", func(t *testing.T) {
		err := dr.Err()
		testy.Error(t, eErr, err)
	})
}
