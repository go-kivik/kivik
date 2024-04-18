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
)

func (d *db) updateDesignDoc(ctx context.Context, tx *sql.Tx, rev revision, data *docData) error {
	if !data.IsDesignDoc() {
		return nil
	}
	stmt, err := tx.PrepareContext(ctx, d.query(`
		INSERT INTO {{ .Design }} (id, rev, rev_id, language, func_type, func_name, func_body, auto_update)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for name, view := range data.DesignFields.Views {
		if view.Map != "" {
			if _, err := stmt.ExecContext(ctx, data.ID, rev.rev, rev.id, data.DesignFields.Language, "map", name, view.Map, data.DesignFields.AutoUpdate); err != nil {
				return err
			}
			if err := d.createViewMap(ctx, tx, data.ID, name, rev.String()); err != nil {
				return err
			}
		}
		if view.Reduce != "" {
			if _, err := stmt.ExecContext(ctx, data.ID, rev.rev, rev.id, data.DesignFields.Language, "reduce", name, view.Reduce, data.DesignFields.AutoUpdate); err != nil {
				return err
			}
		}
	}
	for name, update := range data.DesignFields.Updates {
		if _, err := stmt.ExecContext(ctx, data.ID, rev.rev, rev.id, data.DesignFields.Language, "update", name, update, data.DesignFields.AutoUpdate); err != nil {
			return err
		}
	}
	for name, filter := range data.DesignFields.Filters {
		if _, err := stmt.ExecContext(ctx, data.ID, rev.rev, rev.id, data.DesignFields.Language, "filter", name, filter, data.DesignFields.AutoUpdate); err != nil {
			return err
		}
	}
	if data.DesignFields.ValidateDocUpdates != "" {
		if _, err := stmt.ExecContext(ctx, data.ID, rev.rev, rev.id, data.DesignFields.Language, "validate_doc_update", "validate", data.DesignFields.ValidateDocUpdates, data.DesignFields.AutoUpdate); err != nil {
			return err
		}
	}
	return nil
}

func (d *db) createViewMap(ctx context.Context, tx *sql.Tx, ddoc, name, rev string) error {
	for _, query := range viewMapSchema {
		if _, err := tx.ExecContext(ctx, d.ddocQuery(ddoc, name, rev, query)); err != nil {
			return err
		}
	}
	return nil
}
