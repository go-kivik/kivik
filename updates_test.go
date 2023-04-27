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
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
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
				iter: &iter{closed: true},
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
	testy.Error(t, expected, err)
}

func TestDBUpdatesErr(t *testing.T) {
	expected := "foo error"
	u := &DBUpdates{
		iter: &iter{lasterr: errors.New(expected)},
	}
	err := u.Err()
	testy.Error(t, expected, err)
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
	testy.Error(t, expected, err)
}

func TestDBUpdatesIteratorNew(t *testing.T) {
	u := newDBUpdates(context.Background(), &mock.DBUpdates{})
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
					DBUpdatesFunc: func(_ context.Context) (driver.DBUpdates, error) {
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
					DBUpdatesFunc: func(_ context.Context) (driver.DBUpdates, error) {
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
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.DBUpdates(context.Background())
			testy.StatusError(t, test.err, test.status, err)
			result.cancel = nil // Determinism
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
