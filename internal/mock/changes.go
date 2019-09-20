package mock

import "github.com/go-kivik/kivik/driver"

// Changes mocks driver.Changes
type Changes struct {
	NextFunc    func(*driver.Change) error
	CloseFunc   func() error
	LastSeqFunc func() string
	PendingFunc func() int64
	ETagFunc    func() string
}

var _ driver.Changes = &Changes{}

// Next calls c.NextFunc
func (c *Changes) Next(change *driver.Change) error {
	return c.NextFunc(change)
}

// Close calls c.CloseFunc
func (c *Changes) Close() error {
	return c.CloseFunc()
}

// LastSeq calls c.LastSeqFunc
func (c *Changes) LastSeq() string {
	return c.LastSeqFunc()
}

// Pending calls c.PendingFunc
func (c *Changes) Pending() int64 {
	return c.PendingFunc()
}

// ETag calls c.ETagFunc
func (c *Changes) ETag() string {
	return c.ETagFunc()
}
