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
	"encoding/json"
	"strconv"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/mango"
)

func mangoIndexName(dbName, ddoc, indexName string) string {
	return strconv.Quote("idx_" + tablePrefix + dbName + "$mango_" + md5sumString(ddoc + "/" + indexName)[:8])
}

// GetIndexes returns the list of all indexes in the database.
func (d *db) GetIndexes(ctx context.Context, _ driver.Options) ([]driver.Index, error) {
	indexes := []driver.Index{
		{
			Name:       "_all_docs",
			Type:       "special",
			Definition: map[string]interface{}{"fields": []map[string]string{{"_id": "asc"}}},
		},
	}

	rows, err := d.db.QueryContext(ctx, d.query(`
		SELECT ddoc, name, index_def FROM {{ .MangoIndexes }}
	`))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var ddoc, name, indexDef string
		if err := rows.Scan(&ddoc, &name, &indexDef); err != nil {
			return nil, err
		}

		normalizedFields, err := mango.NormalizeIndexFields(indexDef)
		if err != nil {
			return nil, err
		}

		indexes = append(indexes, driver.Index{
			DesignDoc:  ddoc,
			Name:       name,
			Type:       "json",
			Definition: map[string]interface{}{"fields": normalizedFields},
		})
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return indexes, nil
}

// CreateIndex creates a Mango index.
func (d *db) CreateIndex(ctx context.Context, ddoc, name string, index any, _ driver.Options) error {
	var indexDef []byte
	switch t := index.(type) {
	case json.RawMessage:
		indexDef = t
	case []byte:
		indexDef = t
	case string:
		indexDef = []byte(t)
	default:
		var err error
		indexDef, err = json.Marshal(index)
		if err != nil {
			return err
		}
	}

	_, err := d.db.ExecContext(ctx, d.query(`
		INSERT INTO {{ .MangoIndexes }} (ddoc, name, index_def)
		VALUES ($1, $2, $3)
	`), ddoc, name, string(indexDef))
	return err
}

// DeleteIndex deletes a Mango index.
func (d *db) DeleteIndex(ctx context.Context, ddoc, name string, _ driver.Options) error {
	_, err := d.db.ExecContext(ctx, d.query(`
		DELETE FROM {{ .MangoIndexes }}
		WHERE ddoc = $1 AND name = $2
	`), ddoc, name)
	return err
}
