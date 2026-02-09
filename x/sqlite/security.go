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
	"database/sql"
	"encoding/json"
	"errors"

	"github.com/go-kivik/kivik/v4/driver"
)

// Security returns the database's security document.
func (d *db) Security(ctx context.Context) (*driver.Security, error) {
	var secJSON string
	err := d.db.QueryRowContext(ctx, d.query(`
		SELECT security FROM {{ .Security }} WHERE id = 1
	`)).Scan(&secJSON)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return &driver.Security{}, nil
		}
		return nil, err
	}

	var sec driver.Security
	if err := json.Unmarshal([]byte(secJSON), &sec); err != nil {
		return nil, err
	}
	return &sec, nil
}

// SetSecurity sets the database's security document.
func (d *db) SetSecurity(ctx context.Context, security *driver.Security) error {
	secJSON, err := json.Marshal(security)
	if err != nil {
		return err
	}
	_, err = d.db.ExecContext(ctx, d.query(`
		INSERT OR REPLACE INTO {{ .Security }} (id, security) VALUES (1, ?)
	`), string(secJSON))
	return err
}
