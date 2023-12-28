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

/*
Security Objects

This driver supports fetching and storing security objects, but completely
ignores them for access control. This support is intended only for the purpose
of syncing to/from CouchDB instances.
*/

package fs

import (
	"context"

	"github.com/go-kivik/kivik/v4/driver"
)

func (d *db) Security(ctx context.Context) (*driver.Security, error) {
	return d.cdb.ReadSecurity(ctx, d.path())
}

func (d *db) SetSecurity(ctx context.Context, sec *driver.Security) error {
	return d.cdb.WriteSecurity(ctx, d.path(), sec)
}
