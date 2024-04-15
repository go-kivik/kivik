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
	"errors"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) OpenRevs(ctx context.Context, docID string, revs []string, _ driver.Options) (driver.Rows, error) {
	if len(revs) == 1 && revs[0] == "all" {
		tx, err := d.db.Begin()
		if err != nil {
			return nil, err
		}
		defer tx.Rollback()

		_, _, err = d.winningRev(ctx, tx, docID)
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, &internal.Error{Status: http.StatusNotFound, Message: "missing"}
		case err != nil:
			return nil, err
		}
		return nil, tx.Commit()
	}
	for _, rev := range revs {
		if _, err := parseRev(rev); err != nil {
			return nil, &internal.Error{Message: "invalid rev format", Status: http.StatusBadRequest}
		}
	}
	return nil, nil
}
