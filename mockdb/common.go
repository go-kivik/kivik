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
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

type expectation interface {
	fulfill()
	fulfilled() bool
	Lock()
	Unlock()
	fmt.Stringer
	// method should return the name of the method that would trigger this
	// condition. If verbose is true, the output should disambiguate between
	// different calls to the same method.
	method(verbose bool) string
	error() error
	wait(context.Context) error
	// met is called on the actual value, and returns true if the expectation
	// is met.
	met(expectation) bool
	// for DB methods, this returns the associated *DB object
	dbo() *DB
	opts() kivik.Option
}

// commonExpectation satisfies the expectation interface, except the String()
// and method() methods.
type commonExpectation struct {
	sync.Mutex
	triggered bool
	err       error // nolint: structcheck
	delay     time.Duration
	options   kivik.Option
	db        *DB
}

func (e *commonExpectation) opts() kivik.Option {
	return e.options
}

func (e *commonExpectation) dbo() *DB {
	return e.db
}

func (e *commonExpectation) fulfill() {
	e.triggered = true
}

func (e *commonExpectation) fulfilled() bool {
	return e.triggered
}

func (e *commonExpectation) error() error {
	return e.err
}

// wait blocks until e.delay expires, or ctx is cancelled. If e.delay expires,
// e.err is returned, otherwise ctx.Err() is returned.
func (e *commonExpectation) wait(ctx context.Context) error {
	if e.delay == 0 {
		return e.err
	}
	if err := pause(ctx, e.delay); err != nil {
		return err
	}
	return e.err
}

func pause(ctx context.Context, delay time.Duration) error {
	if delay == 0 {
		return nil
	}
	t := time.NewTimer(delay)
	defer t.Stop()
	select {
	case <-t.C:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

const (
	defaultOptionPlaceholder = "[?]"
)

func formatOptions(o driver.Options) string {
	if o != nil {
		if str := fmt.Sprintf("%v", o); str != "" {
			return str
		}
	}
	return defaultOptionPlaceholder
}

type multiOptions []kivik.Option

var _ kivik.Option = (multiOptions)(nil)

func (mo multiOptions) Apply(t interface{}) {
	if mo == nil {
		return
	}
	for _, opt := range mo {
		if opt != nil {
			opt.Apply(t)
		}
	}
}

func (mo multiOptions) String() string {
	if mo == nil {
		return ""
	}
	parts := make([]string, 0, len(mo))
	for _, o := range mo {
		if o != nil {
			if part := fmt.Sprintf("%s", o); part != "" {
				parts = append(parts, part)
			}
		}
	}
	return strings.Join(parts, ",")
}
