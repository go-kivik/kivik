package mock

import "github.com/go-kivik/kivik/driver"

// Changes mocks driver.Changes
type Changes struct {
	NextFunc  func(*driver.Change) error
	CloseFunc func() error
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

// LastSeq returns the Last Seq of the current result.
func (c *Changes) LastSeq() driver.SequenceID {
	return ""
}

// Pending returns the Pending rows of the current changes.
func (c *Changes) Pending() int64 {
	return 0
}
