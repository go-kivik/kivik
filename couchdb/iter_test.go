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

package couchdb

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"gitlab.com/flimzy/testy"
)

func TestCancelableReadCloser(t *testing.T) {
	t.Run("no cancelation", func(t *testing.T) {
		t.Parallel()
		rc := newCancelableReadCloser(
			context.Background(),
			io.NopCloser(strings.NewReader("foo")),
		)
		result, err := io.ReadAll(rc)
		testy.Error(t, "", err)
		if string(result) != "foo" {
			t.Errorf("Unexpected result: %s", string(result))
		}
	})
	t.Run("pre-canceled", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		rc := newCancelableReadCloser(
			ctx,
			io.NopCloser(strings.NewReader("foo")),
		)
		result, err := io.ReadAll(rc)
		testy.Error(t, "context canceled", err)
		if string(result) != "" {
			t.Errorf("Unexpected result: %s", string(result))
		}
	})
	t.Run("canceled mid-flight", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Millisecond)
		defer cancel()
		r := io.MultiReader(
			strings.NewReader("foo"),
			testy.DelayReader(time.Second),
			strings.NewReader("bar"),
		)
		rc := newCancelableReadCloser(
			ctx,
			io.NopCloser(r),
		)
		result, err := io.ReadAll(rc)
		testy.Error(t, "context deadline exceeded", err)
		if string(result) != "" {
			t.Errorf("Unexpected result: %s", string(result))
		}
	})
	t.Run("read error, not canceled", func(t *testing.T) {
		t.Parallel()
		rc := newCancelableReadCloser(
			context.Background(),
			io.NopCloser(testy.ErrorReader("foo", errors.New("read err"))),
		)
		result, err := io.ReadAll(rc)
		testy.Error(t, "read err", err)
		if string(result) != "" {
			t.Errorf("Unexpected result: %s", string(result))
		}
	})
	t.Run("closed early", func(t *testing.T) {
		t.Parallel()
		rc := newCancelableReadCloser(
			context.Background(),
			io.NopCloser(testy.NeverReader()),
		)
		_ = rc.Close()
		result, err := io.ReadAll(rc)
		testy.Error(t, "iterator closed", err)
		if string(result) != "" {
			t.Errorf("Unexpected result: %s", string(result))
		}
	})
}
