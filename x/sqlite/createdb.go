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
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func (c *client) CreateDB(ctx context.Context, name string, _ driver.Options) error {
	if err := validateDBName(name); err != nil {
		return err
	}

	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	d := c.newDB(name)
	for _, query := range schema {
		_, err := tx.ExecContext(ctx, d.query(query))
		if err != nil {
			if errIsAlreadyExists(err) {
				return &internal.Error{Status: http.StatusPreconditionFailed, Message: "database already exists"}
			}
			return err
		}
	}

	if err := c.logDBUpdate(ctx, tx, name, "created"); err != nil {
		return err
	}

	return tx.Commit()
}
