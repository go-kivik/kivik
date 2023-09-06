package kivikmock

import (
	"context"
	"fmt"
	"sync"
	"time"

	kivik "github.com/go-kivik/kivik/v4"
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
	opts() kivik.Options
}

// commonExpectation satisfies the expectation interface, except the String()
// and method() methods.
type commonExpectation struct {
	sync.Mutex
	triggered bool
	err       error // nolint: structcheck
	delay     time.Duration
	options   kivik.Options
	db        *DB
}

func (e *commonExpectation) opts() kivik.Options {
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
