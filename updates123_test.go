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

//go:build go1.23

package kivik

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestDBUpdatesIterator(t *testing.T) {
	t.Parallel()

	want := []string{"a", "b", "c"}
	var idx int
	updates := newDBUpdates(context.Background(), nil, &mock.DBUpdates{
		NextFunc: func(u *driver.DBUpdate) error {
			if idx >= len(want) {
				return io.EOF
			}
			u.DBName = want[idx]
			idx++
			return nil
		},
	})

	ids := []string{}
	for update, err := range updates.Iterator() {
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		ids = append(ids, update.DBName)
	}

	if diff := cmp.Diff(want, ids); diff != "" {
		t.Errorf("Unexpected updates: %s", diff)
	}
}

func TestDBUpdatesIteratorError(t *testing.T) {
	t.Parallel()

	updates := newDBUpdates(context.Background(), nil, &mock.DBUpdates{
		NextFunc: func(*driver.DBUpdate) error {
			return errors.New("Failure")
		},
	})

	for _, err := range updates.Iterator() {
		if err == nil {
			t.Fatal("Expected error")
		}
		return
	}

	t.Fatal("Expected an error during iteration")
}

func TestDBUpdatesIteratorBreak(t *testing.T) {
	t.Parallel()

	updates := newDBUpdates(context.Background(), nil, &mock.DBUpdates{
		NextFunc: func(*driver.DBUpdate) error {
			return nil
		},
	})

	for _, err := range updates.Iterator() {
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		break
	}

	if updates.iter.state != stateClosed {
		t.Errorf("Expected iterator to be closed")
	}
}
