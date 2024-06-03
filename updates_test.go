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
	"io"
	"net/http"
	"strconv"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestDBUpdatesNext(t *testing.T) {
	tests := []struct {
		name     string
		updates  *DBUpdates
		expected bool
	}{
		{
			name: "nothing more",
			updates: &DBUpdates{
				iter: &iter{state: stateClosed},
			},
			expected: false,
		},
		{
			name: "more",
			updates: &DBUpdates{
				iter: &iter{
					feed: &mockIterator{
						NextFunc: func(_ interface{}) error { return nil },
					},
				},
			},
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.updates.Next()
			if result != test.expected {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

func TestDBUpdatesClose(t *testing.T) {
	expected := "close error"
	u := &DBUpdates{
		iter: &iter{
			feed: &mockIterator{CloseFunc: func() error { return errors.New(expected) }},
		},
	}
	err := u.Close()
	if !testy.ErrorMatches(expected, err) {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestDBUpdatesErr(t *testing.T) {
	expected := "foo error"
	u := &DBUpdates{
		iter: &iter{err: errors.New(expected)},
	}
	err := u.Err()
	if !testy.ErrorMatches(expected, err) {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestDBUpdatesIteratorNext(t *testing.T) {
	expected := "foo error"
	u := &updatesIterator{
		DBUpdates: &mock.DBUpdates{
			NextFunc: func(_ *driver.DBUpdate) error { return errors.New(expected) },
		},
	}
	var i driver.DBUpdate
	err := u.Next(&i)
	if !testy.ErrorMatches(expected, err) {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestDBUpdatesIteratorNew(t *testing.T) {
	u := newDBUpdates(context.Background(), nil, &mock.DBUpdates{})
	expected := &DBUpdates{
		iter: &iter{
			feed: &updatesIterator{
				DBUpdates: &mock.DBUpdates{},
			},
			curVal: &driver.DBUpdate{},
		},
	}
	u.cancel = nil // determinism
	if d := testy.DiffInterface(expected, u); d != nil {
		t.Error(d)
	}
}

func TestDBUpdateGetters(t *testing.T) {
	dbname := "foo"
	updateType := "chicken"
	seq := "abc123"
	u := &DBUpdates{
		iter: &iter{
			state: stateRowReady,
			curVal: &driver.DBUpdate{
				DBName: dbname,
				Type:   updateType,
				Seq:    seq,
			},
		},
	}

	t.Run("DBName", func(t *testing.T) {
		result := u.DBName()
		if result != dbname {
			t.Errorf("Unexpected result: %s", result)
		}
	})

	t.Run("Type", func(t *testing.T) {
		result := u.Type()
		if result != updateType {
			t.Errorf("Unexpected result: %s", result)
		}
	})

	t.Run("Seq", func(t *testing.T) {
		result := u.Seq()
		if result != seq {
			t.Errorf("Unexpected result: %s", result)
		}
	})

	t.Run("LastSeq, should error during iteration", func(t *testing.T) {
		result, err := u.LastSeq()
		if result != "" {
			t.Errorf("Unexpected result: %s", result)
		}
		if !testy.ErrorMatches("LastSeq must not be called until results iteration is complete", err) {
			t.Errorf("Unexpected error: %s", err)
		}
	})

	t.Run("Not Ready", func(t *testing.T) {
		u.state = stateReady

		t.Run("DBName", func(t *testing.T) {
			result := u.DBName()
			if result != "" {
				t.Errorf("Unexpected result: %s", result)
			}
		})

		t.Run("Type", func(t *testing.T) {
			result := u.Type()
			if result != "" {
				t.Errorf("Unexpected result: %s", result)
			}
		})

		t.Run("Seq", func(t *testing.T) {
			result := u.Seq()
			if result != "" {
				t.Errorf("Unexpected result: %s", result)
			}
		})
	})
}

func TestDBUpdates(t *testing.T) {
	tests := []struct {
		name     string
		client   *Client
		expected *DBUpdates
		status   int
		err      string
	}{
		{
			name: "non-DBUpdater",
			client: &Client{
				driverClient: &mock.Client{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not implement DBUpdater",
		},
		{
			name: "db error",
			client: &Client{
				driverClient: &mock.DBUpdater{
					DBUpdatesFunc: func(context.Context, driver.Options) (driver.DBUpdates, error) {
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
				driverClient: &mock.DBUpdater{
					DBUpdatesFunc: func(context.Context, driver.Options) (driver.DBUpdates, error) {
						return &mock.DBUpdates{ID: "a"}, nil
					},
				},
			},
			expected: &DBUpdates{
				iter: &iter{
					feed: &updatesIterator{
						DBUpdates: &mock.DBUpdates{ID: "a"},
					},
					curVal: &driver.DBUpdate{},
				},
			},
		},
		{
			name: "client closed",
			client: &Client{
				closed:       true,
				driverClient: &mock.DBUpdater{},
			},
			status: http.StatusServiceUnavailable,
			err:    "kivik: client closed",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.client.DBUpdates(context.Background())
			err := result.Err()
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if err != nil {
				return
			}
			result.cancel = nil  // Determinism
			result.onClose = nil // Determinism
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
	t.Run("standalone", func(t *testing.T) {
		t.Run("after err, close doesn't block", func(t *testing.T) {
			client := &Client{
				driverClient: &mock.DBUpdater{
					DBUpdatesFunc: func(context.Context, driver.Options) (driver.DBUpdates, error) {
						return nil, errors.New("asdfsad")
					},
				},
			}
			rows := client.DBUpdates(context.Background())
			if err := rows.Err(); err == nil {
				t.Fatal("expected an error, got none")
			}
			_ = client.Close() // Should not block
		})
		t.Run("not updater, close doesn't block", func(t *testing.T) {
			client := &Client{
				driverClient: &mock.Client{},
			}
			rows := client.DBUpdates(context.Background())
			if err := rows.Err(); err == nil {
				t.Fatal("expected an error, got none")
			}
			_ = client.Close() // Should not block
		})
	})
}

func TestDBUpdates_Next_resets_iterator_value(t *testing.T) {
	idx := 0
	client := &Client{
		driverClient: &mock.DBUpdater{
			DBUpdatesFunc: func(context.Context, driver.Options) (driver.DBUpdates, error) {
				return &mock.DBUpdates{
					NextFunc: func(update *driver.DBUpdate) error {
						idx++
						switch idx {
						case 1:
							update.DBName = strconv.Itoa(idx)
							return nil
						case 2:
							return nil
						}
						return io.EOF
					},
				}, nil
			},
		},
	}

	updates := client.DBUpdates(context.Background())

	wantDBNames := []string{"1", ""}
	gotDBNames := []string{}
	for updates.Next() {
		gotDBNames = append(gotDBNames, updates.DBName())
	}
	if d := cmp.Diff(wantDBNames, gotDBNames); d != "" {
		t.Error(d)
	}
}

func TestDBUpdates_LastSeq(t *testing.T) {
	t.Run("non-LastSeqer", func(t *testing.T) {
		client := &Client{
			driverClient: &mock.DBUpdater{
				DBUpdatesFunc: func(context.Context, driver.Options) (driver.DBUpdates, error) {
					return &mock.DBUpdates{
						NextFunc: func(_ *driver.DBUpdate) error {
							return io.EOF
						},
					}, nil
				},
			},
		}

		updates := client.DBUpdates(context.Background())
		for updates.Next() {
			/* .. do nothing .. */
		}
		lastSeq, err := updates.LastSeq()
		if err != nil {
			t.Fatal(err)
		}
		if lastSeq != "" {
			t.Errorf("Unexpected lastSeq: %s", lastSeq)
		}
	})
	t.Run("LastSeqer", func(t *testing.T) {
		client := &Client{
			driverClient: &mock.DBUpdater{
				DBUpdatesFunc: func(context.Context, driver.Options) (driver.DBUpdates, error) {
					return &mock.LastSeqer{
						DBUpdates: &mock.DBUpdates{
							NextFunc: func(_ *driver.DBUpdate) error {
								return io.EOF
							},
						},
						LastSeqFunc: func() (string, error) {
							return "99-last", nil
						},
					}, nil
				},
			},
		}

		updates := client.DBUpdates(context.Background())
		for updates.Next() {
			/* .. do nothing .. */
		}
		lastSeq, err := updates.LastSeq()
		if err != nil {
			t.Fatal(err)
		}
		if lastSeq != "99-last" {
			t.Errorf("Unexpected lastSeq: %s", lastSeq)
		}
	})
}
