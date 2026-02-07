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
	"fmt"
	"io"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"
)

type TestFeed struct {
	max      int64
	i        int64
	closeErr error
}

var _ iterator = &TestFeed{}

func (f *TestFeed) Close() error { return f.closeErr }
func (f *TestFeed) Next(ifce interface{}) error {
	i, ok := ifce.(*int64)
	if ok {
		*i = f.i
		f.i++
		if f.i > f.max {
			return io.EOF
		}
		time.Sleep(5 * time.Millisecond)
		return nil
	}
	panic(fmt.Sprintf("unknown type: %T", ifce))
}

func TestIterator(t *testing.T) {
	iter := newIterator(context.Background(), nil, &TestFeed{max: 10}, func() interface{} { var i int64; return &i }())
	expected := []int64{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	result := []int64{}
	for iter.Next() {
		val, ok := iter.curVal.(*int64)
		if !ok {
			panic("Unexpected type")
		}
		result = append(result, *val)
	}
	if err := iter.Err(); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if d := testy.DiffAsJSON(expected, result); d != nil {
		t.Errorf("Unexpected result:\n%s\n", d)
	}
}

func TestCancelledIterator(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	iter := newIterator(ctx, nil, &TestFeed{max: 10000}, func() interface{} { var i int64; return &i }())
	for iter.Next() { //nolint:revive // empty block necessary for loop
	}
	if err := iter.Err(); err.Error() != "context deadline exceeded" {
		t.Errorf("Unexpected error: %s", err)
	}
}

// blockingFeed is a feed whose Next blocks until Close is called,
// simulating a continuous changes feed waiting on network I/O.
type blockingFeed struct {
	ready chan struct{}
	done  chan struct{}
}

var _ iterator = &blockingFeed{}

func (f *blockingFeed) Close() error {
	select {
	case <-f.done:
	default:
		close(f.done)
	}
	return nil
}

func (f *blockingFeed) Next(_ interface{}) error {
	f.ready <- struct{}{}
	<-f.done
	return context.Canceled
}

func TestCloseWhileNextBlocked(t *testing.T) {
	t.Parallel()

	feed := &blockingFeed{
		ready: make(chan struct{}),
		done:  make(chan struct{}),
	}
	it := newIterator(context.Background(), nil, feed, new(int64))

	go it.Next()

	<-feed.ready

	closeDone := make(chan struct{})
	go func() {
		_ = it.Close()
		close(closeDone)
	}()

	select {
	case <-closeDone:
	case <-time.After(5 * time.Second):
		t.Fatal("iter.Close deadlocked while Next was blocked")
	}

	if err := it.Err(); err != nil {
		t.Errorf("Err() after Close should be nil, got: %s", err)
	}
}

// TestCloseErrSuppressesErrorWhenClosing verifies that closeErr does not
// store an error when the iterator is being closed by Close. This covers the
// awaitDone path where closeErr(context.Canceled) is called after Close
// cancels the context.
func TestCloseErrSuppressesErrorWhenClosing(t *testing.T) {
	t.Parallel()

	it := &iter{
		feed:  &TestFeed{},
		state: stateRowReady,
	}
	it.closing.Store(true)
	_ = it.closeErr(context.Canceled)

	if it.err != nil {
		t.Errorf("expected nil err when closing, got: %s", it.err)
	}
}

func Test_iter_isReady(t *testing.T) {
	tests := []struct {
		name string
		iter *iter
		err  string
	}{
		{
			name: "not ready",
			iter: &iter{},
			err:  "kivik: Iterator access before calling Next",
		},
		{
			name: "closed",
			iter: &iter{state: stateClosed},
			err:  "kivik: Iterator is closed",
		},
		{
			name: "success",
			iter: &iter{state: stateRowReady},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.iter.isReady()
			if !testy.ErrorMatches(test.err, err) {
				t.Errorf("Unexpected error: %s", err)
			}
		})
	}
}
