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

package sqlite

import (
	"context"
	"io"

	"github.com/go-kivik/kivik/v4/driver"
)

type changes struct{}

var _ driver.Changes = &changes{}

func (c *changes) Next(change *driver.Change) error {
	return io.EOF
}

func (c *changes) Close() error {
	return nil
}

func (c *changes) LastSeq() string {
	return ""
}

func (c *changes) Pending() int64 {
	return 0
}

func (c *changes) ETag() string {
	return ""
}

func (db) Changes(context.Context, driver.Options) (driver.Changes, error) {
	return &changes{}, nil
}
