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

package memorydb

import (
	"context"
	"errors"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
)

func cloneSecurity(in *driver.Security) *driver.Security {
	return &driver.Security{
		Admins: driver.Members{
			Names: in.Admins.Names,
			Roles: in.Admins.Roles,
		},
		Members: driver.Members{
			Names: in.Members.Names,
			Roles: in.Members.Roles,
		},
	}
}

func (d *db) Security(ctx context.Context) (*driver.Security, error) {
	if exists, _ := d.DBExists(ctx, d.dbName, nil); !exists {
		return nil, statusError{status: http.StatusNotFound, error: errors.New("database does not exist")}
	}
	d.db.mu.RLock()
	defer d.db.mu.RUnlock()
	if d.db.deleted {
		return nil, statusError{status: http.StatusNotFound, error: errors.New("missing")}
	}
	return cloneSecurity(d.db.security), nil
}

func (d *db) SetSecurity(_ context.Context, sec *driver.Security) error {
	d.db.mu.Lock()
	defer d.db.mu.Unlock()
	if d.db.deleted {
		return statusError{status: http.StatusNotFound, error: errors.New("missing")}
	}
	d.db.security = cloneSecurity(sec)
	return nil
}
