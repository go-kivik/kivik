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
	"net/http"

	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func (d *db) updateDesignDoc(ctx context.Context, tx *sql.Tx, rev revision, curRev revision, data *docData) error {
	if !data.IsDesignDoc() {
		return nil
	}
	if err := d.dropMapTables(ctx, tx, data.ID, curRev); err != nil {
		return err
	}

	stmt, err := tx.PrepareContext(ctx, d.query(`
		INSERT INTO {{ .Design }} (id, rev, rev_id, language, func_type, func_name, func_body, auto_update, include_design, collation, local_seq)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`))
	if err != nil {
		return err
	}
	defer stmt.Close()

	for name, view := range data.DesignFields.Views {
		if view.Map != "" {
			if _, err := stmt.ExecContext(ctx,
				data.ID, rev.rev, rev.id, data.DesignFields.Language, "map", name, view.Map,
				data.DesignFields.AutoUpdate, data.DesignFields.Options.IncludeDesign, data.DesignFields.Options.Collation, data.DesignFields.Options.LocalSeq,
			); err != nil {
				if errIsInvalidCollation(err) {
					return &internal.Error{Status: http.StatusBadRequest, Message: "unsupported collation: " + *data.DesignFields.Options.Collation}
				}
				return err
			}
			if err := d.createViewMap(ctx, tx, data.ID, name, rev.String(), data.DesignFields.Options.Collation); err != nil {
				return err
			}
		}
		if view.Reduce != "" {
			if _, err := stmt.ExecContext(ctx, data.ID, rev.rev, rev.id, data.DesignFields.Language, "reduce", name, view.Reduce, data.DesignFields.AutoUpdate, nil, nil, nil); err != nil {
				return err
			}
		}
	}
	for name, update := range data.DesignFields.Updates {
		if _, err := stmt.ExecContext(ctx, data.ID, rev.rev, rev.id, data.DesignFields.Language, "update", name, update, data.DesignFields.AutoUpdate, nil, nil, nil); err != nil {
			return err
		}
	}
	for name, filter := range data.DesignFields.Filters {
		if _, err := stmt.ExecContext(ctx, data.ID, rev.rev, rev.id, data.DesignFields.Language, "filter", name, filter, data.DesignFields.AutoUpdate, nil, nil, nil); err != nil {
			return err
		}
	}
	if data.DesignFields.ValidateDocUpdates != "" {
		if _, err := stmt.ExecContext(ctx, data.ID, rev.rev, rev.id, data.DesignFields.Language, "validate", "validate", data.DesignFields.ValidateDocUpdates, data.DesignFields.AutoUpdate, nil, nil, nil); err != nil {
			return err
		}
	}
	return nil
}

func (d *db) dropMapTables(ctx context.Context, tx *sql.Tx, docID string, curRev revision) error {
	if curRev.rev == 0 {
		return nil
	}
	rows, err := tx.QueryContext(ctx, d.query(`
		SELECT id, rev, rev_id, func_name
		FROM {{ .Design }}
		WHERE func_type = 'map' AND id = $1
			AND rev = $2 AND rev_id = $3
	`), docID, curRev.rev, curRev.id)
	if err != nil {
		return err
	}
	defer rows.Close()

	var queries []string
	for rows.Next() {
		var (
			id, view string
			rev      revision
		)
		if err := rows.Scan(&id, &rev.rev, &rev.id, &view); err != nil {
			return err
		}
		queries = append(queries, d.ddocQuery(id, view, rev.String(), `DROP TABLE IF EXISTS {{ .Map }}`))
	}
	if err := rows.Err(); err != nil {
		return err
	}

	for _, q := range queries {
		if _, err := tx.ExecContext(ctx, q); err != nil {
			return err
		}
	}
	return nil
}

func (d *db) createViewMap(ctx context.Context, tx *sql.Tx, ddoc, name, rev string, collation *string) error {
	for _, query := range viewSchema {
		if _, err := tx.ExecContext(ctx, d.createDdocQuery(ddoc, name, rev, query, collation)); err != nil {
			return err
		}
	}
	return nil
}
