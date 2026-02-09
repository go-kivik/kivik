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

func (c *canceledContext) Value(_ any) any { return nil }
