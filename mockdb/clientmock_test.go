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

package mockdb

import (
	"errors"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
)

func TestReplication(t *testing.T) {
	_, m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	ts := time.Now()
	eID := "a1"
	eSource := "a2"
	eTarget := "a3"
	eStartTime := ts
	eEndTime := ts.Add(time.Second)
	eState := kivik.ReplicationComplete
	eErr := "a5"
	r := m.NewReplication().
		ID(eID).
		Source(eSource).
		Target(eTarget).
		StartTime(eStartTime).
		EndTime(eEndTime).
		State(eState).
		Err(errors.New(eErr))
	dr := &driverReplication{r}
	t.Run("ID", func(t *testing.T) {
		if id := dr.ReplicationID(); id != eID {
			t.Errorf("Unexpected ID. Got %s, want %s", id, eID)
		}
	})
	t.Run("Source", func(t *testing.T) {
		if s := dr.Source(); s != eSource {
			t.Errorf("Unexpected Source. Got %s, want %s", s, eSource)
		}
	})
	t.Run("StartTime", func(t *testing.T) {
		if ts := dr.StartTime(); !ts.Equal(eStartTime) {
			t.Errorf("Unexpected StartTime. Got %s, want %s", ts, eStartTime)
		}
	})
	t.Run("EndTime", func(t *testing.T) {
		if ts := dr.EndTime(); !ts.Equal(eEndTime) {
			t.Errorf("Unexpected EndTime. Got %s, want %s", ts, eEndTime)
		}
	})
	t.Run("State", func(t *testing.T) {
		if s := kivik.ReplicationState(dr.State()); s != eState {
			t.Errorf("Unexpected State. Got %s, want %s", s, eState)
		}
	})
	t.Run("Err", func(t *testing.T) {
		err := dr.Err()
		if !testy.ErrorMatches(eErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}
