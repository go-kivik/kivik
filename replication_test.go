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

package kivik

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

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

func TestDocsRead(t *testing.T) {
	t.Run("No Info", func(t *testing.T) {
		r := &Replication{}
		result := r.DocsRead()
		if result != 0 {
			t.Errorf("Unexpected doc count: %d", result)
		}
	})
	t.Run("With Info", func(t *testing.T) {
		r := &Replication{
			info: &driver.ReplicationInfo{
				DocsRead: 123,
			},
		}
		result := r.DocsRead()
		if result != 123 {
			t.Errorf("Unexpected doc count: %d", result)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		var r *Replication
		result := r.DocsRead()
		if result != 0 {
			t.Errorf("Unexpected doc count: %d", result)
		}
	})
}

func TestDocWriteFailures(t *testing.T) {
	t.Run("No Info", func(t *testing.T) {
		r := &Replication{}
		result := r.DocWriteFailures()
		if result != 0 {
			t.Errorf("Unexpected doc count: %d", result)
		}
	})
	t.Run("With Info", func(t *testing.T) {
		r := &Replication{
			info: &driver.ReplicationInfo{
				DocWriteFailures: 123,
			},
		}
		result := r.DocWriteFailures()
		if result != 123 {
			t.Errorf("Unexpected doc count: %d", result)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		var r *Replication
		result := r.DocWriteFailures()
		if result != 0 {
			t.Errorf("Unexpected doc count: %d", result)
		}
	})
}

func TestProgress(t *testing.T) {
	t.Run("No Info", func(t *testing.T) {
		r := &Replication{}
		result := r.Progress()
		if result != 0 {
			t.Errorf("Unexpected doc count: %v", result)
		}
	})
	t.Run("With Info", func(t *testing.T) {
		r := &Replication{
			info: &driver.ReplicationInfo{
				Progress: 123,
			},
		}
		result := r.Progress()
		if result != 123 {
			t.Errorf("Unexpected doc count: %v", result)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		var r *Replication
		result := r.Progress()
		if result != 0 {
			t.Errorf("Unexpected doc count: %v", result)
		}
	})
}

func TestNewReplication(t *testing.T) {
	source := "foo"
	target := "bar"
	rep := &mock.Replication{
		SourceFunc: func() string { return source },
		TargetFunc: func() string { return target },
	}
	expected := &Replication{
		Source: source,
		Target: target,
		irep:   rep,
	}
	result := newReplication(rep)
	if d := testy.DiffInterface(expected, result); d != nil {
		t.Error(d)
	}
}

func TestReplicationGetters(t *testing.T) {
	repID := "repID"
	start := parseTime(t, "2018-01-01T00:00:00Z")
	end := parseTime(t, "2019-01-01T00:00:00Z")
	state := "confusion"
	r := newReplication(&mock.Replication{
		MetadataFunc: func() driver.ReplicationMetadata {
			return driver.ReplicationMetadata{
				ID: repID,
			}
		},
		StartTimeFunc: func() time.Time { return start },
		EndTimeFunc:   func() time.Time { return end },
		StateFunc:     func() string { return state },
	})

	t.Run("ReplicationID", func(t *testing.T) {
		result := r.ReplicationID()
		if result != repID {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("StartTime", func(t *testing.T) {
		result := r.StartTime()
		if !result.Equal(start) {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("EndTime", func(t *testing.T) {
		result := r.EndTime()
		if !result.Equal(end) {
			t.Errorf("Unexpected result: %v", result)
		}
	})

	t.Run("State", func(t *testing.T) {
		result := r.State()
		if result != ReplicationState(state) {
			t.Errorf("Unexpected result: %v", result)
		}
	})
}

func TestReplicationErr(t *testing.T) {
	t.Run("No error", func(t *testing.T) {
		r := &Replication{
			irep: &mock.Replication{
				ErrFunc: func() error { return nil },
			},
		}
		if err := r.Err(); err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	})
	t.Run("Error", func(t *testing.T) {
		r := &Replication{
			irep: &mock.Replication{
				ErrFunc: func() error {
					return errors.New("rep error")
				},
			},
		}
		if err := r.Err(); err == nil || err.Error() != "rep error" {
			t.Errorf("Unexpected error: %s", err)
		}
	})
	t.Run("Nil", func(t *testing.T) {
		var r *Replication
		if err := r.Err(); err != nil {
			t.Errorf("Unexpected error: %s", err)
		}
	})
}

func TestReplicationIsActive(t *testing.T) {
	t.Run("Active", func(t *testing.T) {
		r := &Replication{
			irep: &mock.Replication{
				StateFunc: func() string {
					return "active"
				},
			},
		}
		if !r.IsActive() {
			t.Errorf("Expected active")
		}
	})
	t.Run("Complete", func(t *testing.T) {
		r := &Replication{
			irep: &mock.Replication{
				StateFunc: func() string {
					return string(ReplicationComplete)
				},
			},
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

func TestReplicationDelete(t *testing.T) {
	expected := "delete error"
	r := &Replication{
		irep: &mock.Replication{
			DeleteFunc: func(context.Context) error { return errors.New(expected) },
		},
	}
	err := r.Delete(context.Background())
	testy.Error(t, expected, err)
}

func TestReplicationUpdate(t *testing.T) {
	t.Run("update error", func(t *testing.T) {
		expected := "rep error"
		r := &Replication{
			irep: &mock.Replication{
				UpdateFunc: func(context.Context, *driver.ReplicationInfo) error {
					return errors.New(expected)
				},
			},
		}
		err := r.Update(context.Background())
		testy.Error(t, expected, err)
	})

	t.Run("success", func(t *testing.T) {
		expected := driver.ReplicationInfo{
			DocsRead: 123,
		}
		r := &Replication{
			irep: &mock.Replication{
				UpdateFunc: func(_ context.Context, i *driver.ReplicationInfo) error {
					*i = driver.ReplicationInfo{
						DocsRead: 123,
					}
					return nil
				},
			},
		}
		err := r.Update(context.Background())
		testy.Error(t, "", err)
		if d := testy.DiffInterface(&expected, r.info); d != nil {
			t.Error(d)
		}
	})
}

func TestGetReplications(t *testing.T) {
	tests := []struct {
		name     string
		client   *Client
		options  Option
		expected []*Replication
		status   int
		err      string
	}{
		{
			name: "non-replicator",
			client: &Client{
				driverClient: &mock.Client{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support replication",
		},
		{
			name: "db error",
			client: &Client{
				driverClient: &mock.ClientReplicator{
					GetReplicationsFunc: func(context.Context, driver.Options) ([]driver.Replication, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			client: &Client{
				driverClient: &mock.ClientReplicator{
					GetReplicationsFunc: func(_ context.Context, options driver.Options) ([]driver.Replication, error) {
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						wantOpts := map[string]interface{}{"foo": 123}
						if d := testy.DiffInterface(wantOpts, gotOpts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%v", d)
						}
						return []driver.Replication{
							&mock.Replication{ID: "1"},
							&mock.Replication{ID: "2"},
						}, nil
					},
				},
			},
			options: Param("foo", 123),
			expected: []*Replication{
				{
					Source: "1-source",
					Target: "1-target",
					irep:   &mock.Replication{ID: "1"},
				},
				{
					Source: "2-source",
					Target: "2-target",
					irep:   &mock.Replication{ID: "2"},
				},
			},
		},
		{
			name: "closed",
			client: &Client{
				closed: 1,
			},
			status: http.StatusServiceUnavailable,
			err:    "client closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.GetReplications(context.Background(), test.options)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestReplicate(t *testing.T) {
	tests := []struct {
		name           string
		client         *Client
		target, source string
		options        Option
		expected       *Replication
		status         int
		err            string
	}{
		{
			name: "non-replicator",
			client: &Client{
				driverClient: &mock.Client{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support replication",
		},
		{
			name: "db error",
			client: &Client{
				driverClient: &mock.ClientReplicator{
					ReplicateFunc: func(context.Context, string, string, driver.Options) (driver.Replication, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			client: &Client{
				driverClient: &mock.ClientReplicator{
					ReplicateFunc: func(_ context.Context, target, source string, options driver.Options) (driver.Replication, error) {
						expectedTarget := "foo"
						expectedSource := "bar"
						gotOpts := map[string]interface{}{}
						options.Apply(gotOpts)
						wantOpts := map[string]interface{}{"foo": 123}
						if target != expectedTarget {
							return nil, fmt.Errorf("Unexpected target: %s", target)
						}
						if source != expectedSource {
							return nil, fmt.Errorf("Unexpected source: %s", source)
						}
						if d := testy.DiffInterface(wantOpts, gotOpts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%v", d)
						}
						return &mock.Replication{ID: "a"}, nil
					},
				},
			},
			target:  "foo",
			source:  "bar",
			options: Param("foo", 123),
			expected: &Replication{
				Source: "a-source",
				Target: "a-target",
				irep:   &mock.Replication{ID: "a"},
			},
		},
		{
			name: "closed",
			client: &Client{
				closed: 1,
			},
			status: http.StatusServiceUnavailable,
			err:    "client closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.Replicate(context.Background(), test.target, test.source, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
