package kivikmock

import (
	"context"
	"time"
)

var startime = time.Now()

// canceledContext is immediately canceled
type canceledContext struct {
	ch <-chan struct{}
}

var _ context.Context = &canceledContext{}

func newCanceledContext() context.Context {
	ch := make(chan struct{})
	close(ch)
	return &canceledContext{ch}
}

func (c *canceledContext) Deadline() (time.Time, bool) {
	return startime, true
}

func (c *canceledContext) Done() <-chan struct{} {
	return c.ch
}

func (c *canceledContext) Err() error { return context.Canceled }

func (c *canceledContext) Value(_ interface{}) interface{} { return nil }
