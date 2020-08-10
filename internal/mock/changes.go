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

package mock

import "github.com/go-kivik/kivik/v4/driver"

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
