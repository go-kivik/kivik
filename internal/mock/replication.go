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

import (
	"context"

	"github.com/go-kivik/kivik/v4/driver"
)

// Replication mocks driver.Replication
type Replication struct {
	// ID identifies a specific Replication instance
	ID           string
	DeleteFunc   func(context.Context) error
	MetadataFunc func() driver.ReplicationMetadata
	ErrFunc      func() error
	StateFunc    func() string
	UpdateFunc   func(context.Context, *driver.ReplicationInfo) error
}

var _ driver.Replication = &Replication{}

// Delete calls r.DeleteFunc
func (r *Replication) Delete(ctx context.Context) error {
	return r.DeleteFunc(ctx)
}

// Metadata calls r.MetadataFunc if it is not nil, or returns a default value.
func (r *Replication) Metadata() driver.ReplicationMetadata {
	if r.MetadataFunc != nil {
		return r.MetadataFunc()
	}
	return driver.ReplicationMetadata{
		Source: r.ID + "-source",
		Target: r.ID + "-target",
	}
}

// Err calls r.ErrFunc
func (r *Replication) Err() error {
	return r.ErrFunc()
}

// State calls r.StateFunc
func (r *Replication) State() string {
	return r.StateFunc()
}

// Update calls r.UpdateFunc
func (r *Replication) Update(ctx context.Context, rep *driver.ReplicationInfo) error {
	return r.UpdateFunc(ctx, rep)
}
