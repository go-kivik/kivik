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
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/mango"
	"github.com/go-kivik/kivik/v4/x/options"
)

func (d *db) Find(ctx context.Context, query any, _ driver.Options) (driver.Rows, error) {
	vopts, err := options.FindOptions(query)
	if err != nil {
		return nil, err
	}

	var sortOrderBy string
	if sortFields := vopts.SortFields(); len(sortFields) > 0 {
		sortOrderBy, err = d.sortOrderByFromIndex(ctx, sortFields)
		if err != nil {
			return nil, err
		}
	}

	if ddoc := vopts.UseIndexDdoc(); ddoc != "" {
		var count int
		indexName := vopts.UseIndexName()
		if indexName != "" {
			if err := d.db.QueryRowContext(ctx, d.query(`
				SELECT COUNT(*) FROM {{ .MangoIndexes }} WHERE ddoc = $1 AND name = $2
			`), ddoc, indexName).Scan(&count); err != nil {
				return nil, err
			}
		} else {
			if err := d.db.QueryRowContext(ctx, d.query(`
				SELECT COUNT(*) FROM {{ .MangoIndexes }} WHERE ddoc = $1
			`), ddoc).Scan(&count); err != nil {
				return nil, err
			}
		}
		if count == 0 {
			return nil, &internal.Error{Status: http.StatusBadRequest, Message: fmt.Sprintf("index %q not found", ddoc)}
		}
	}

	var selector json.RawMessage
	input := query.(json.RawMessage)
	var raw struct {
		Selector json.RawMessage `json:"selector"`
	}
	if err := json.Unmarshal(input, &raw); err == nil {
		selector = raw.Selector
	}

	return d.queryBuiltinView(ctx, vopts, selector, sortOrderBy)
}

func (d *db) sortOrderByFromIndex(ctx context.Context, sortFields []options.SortField) (string, error) {
	rows, err := d.db.QueryContext(ctx, d.query(`
		SELECT index_def FROM {{ .MangoIndexes }}
	`))
	if err != nil {
		return "", err
	}
	defer rows.Close()

	for rows.Next() {
		var indexDef string
		if err := rows.Scan(&indexDef); err != nil {
			return "", err
		}
		idxFields, err := mango.ExtractIndexFields([]byte(indexDef))
		if err != nil {
			continue
		}
		if coversSort(idxFields, sortFields) {
			parts := make([]string, len(sortFields))
			for i, sf := range sortFields {
				dir := "ASC"
				if sf.Desc {
					dir = "DESC"
				}
				parts[i] = "json_extract(view.doc, '" + mango.FieldToJSONPath(sf.Field) + "') " + dir
			}
			return "ORDER BY " + strings.Join(parts, ", "), nil
		}
	}
	if err := rows.Err(); err != nil {
		return "", err
	}

	return "", &internal.Error{Status: http.StatusBadRequest, Message: "no index exists for this sort, try indexing by the sort fields"}
}

func coversSort(indexFields []string, sortFields []options.SortField) bool {
	if len(sortFields) > len(indexFields) {
		return false
	}
	for i, sf := range sortFields {
		if indexFields[i] != sf.Field || sf.Desc != sortFields[0].Desc {
			return false
		}
	}
	return true
}

// selectorToSQL translates a Mango selector JSON object into parameterized SQL
// WHERE conditions. It returns condition strings using json_extract(doc.doc, ...)
// expressions and corresponding bind values. argOffset sets the starting $N
// placeholder number so conditions can be appended to an existing argument list.
// Unsupported operators are silently skipped, broadening the result set for the
// in-memory selector.Match() safety net to correct.
func selectorToSQL(selector json.RawMessage, argOffset int) ([]string, []any) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(selector, &fields); err != nil {
		return nil, nil
	}

	var conds []string
	var args []any

	for _, key := range sortedKeys(fields) {
		val := fields[key]
		switch {
		case key == "$and":
			c, a := combineSelectors(val, " AND ", false, argOffset+len(args))
			conds = append(conds, c...)
			args = append(args, a...)

		case key == "$or":
			c, a := combineSelectors(val, " OR ", true, argOffset+len(args))
			conds = append(conds, c...)
			args = append(args, a...)

		case strings.HasPrefix(key, "$"):
			continue

		default:
			jsonPath := mango.FieldToJSONPath(key)
			c, a := fieldCondition(jsonPath, val, argOffset+len(args))
			if c != "" {
				conds = append(conds, c)
				args = append(args, a...)
			}
		}
	}

	if len(conds) == 0 {
		return nil, nil
	}
	return conds, args
}

// combineSelectors unmarshals val as an array of sub-selectors, converts each
// to SQL, and joins them with sep. If wrap is true, the result is parenthesized.
func combineSelectors(val json.RawMessage, sep string, wrap bool, argOffset int) ([]string, []any) {
	var elements []json.RawMessage
	if err := json.Unmarshal(val, &elements); err != nil {
		return nil, nil
	}
	var parts []string
	var args []any
	for _, elem := range elements {
		subConds, subArgs := selectorToSQL(elem, argOffset+len(args))
		parts = append(parts, subConds...)
		args = append(args, subArgs...)
	}
	if len(parts) == 0 {
		return nil, nil
	}
	joined := strings.Join(parts, sep)
	if wrap {
		joined = "(" + joined + ")"
	}
	return []string{joined}, args
}

func fieldCondition(jsonPath string, val json.RawMessage, argOffset int) (string, []any) {
	expr := "json_extract(doc.doc, '" + jsonPath + "')"
	if val[0] != '{' {
		return comparisonCondition(expr, "=", val, argOffset)
	}

	var ops map[string]json.RawMessage
	if err := json.Unmarshal(val, &ops); err != nil {
		return "", nil
	}

	for _, op := range sortedKeys(ops) {
		opVal := ops[op]
		switch op {
		case "$exists":
			var exists bool
			if err := json.Unmarshal(opVal, &exists); err != nil {
				return "", nil
			}
			if exists {
				return expr + " IS NOT NULL", nil
			}
			return expr + " IS NULL", nil

		case "$in":
			var values []json.RawMessage
			if err := json.Unmarshal(opVal, &values); err != nil {
				return "", nil
			}
			args := make([]any, len(values))
			for i, v := range values {
				args[i] = decodeValue(v)
			}
			return expr + " IN (" + placeholders(argOffset+1, len(values)) + ")", args

		case "$gt", "$gte", "$lt", "$lte":
			return inequalityCondition(expr, jsonPath, op, opVal, argOffset)

		case "$eq":
			return comparisonCondition(expr, "=", opVal, argOffset)
		case "$ne":
			return comparisonCondition(expr, "!=", opVal, argOffset)

		default:
			return "", nil
		}
	}

	return "", nil
}

func inequalityCondition(expr, jsonPath, op string, val json.RawMessage, argOffset int) (string, []any) {
	var sqlOp string
	switch op {
	case "$gt":
		sqlOp = ">"
	case "$gte":
		sqlOp = ">="
	case "$lt":
		sqlOp = "<"
	case "$lte":
		sqlOp = "<="
	}

	decoded := decodeValue(val)
	typeExpr := "json_type(doc.doc, '" + jsonPath + "')"

	var typeGuard string
	switch decoded.(type) {
	case float64:
		typeGuard = typeExpr + " NOT IN ('integer', 'real')"
	case string:
		typeGuard = typeExpr + " != 'text'"
	default:
		return "", nil
	}

	placeholder := fmt.Sprintf("$%d", argOffset+1)
	return typeGuard + " OR " + expr + " " + sqlOp + " " + placeholder, []any{decoded}
}

func comparisonCondition(expr, op string, val json.RawMessage, argOffset int) (string, []any) {
	decoded := decodeValue(val)
	if decoded == nil {
		if op == "=" {
			return expr + " IS NULL", nil
		}
		if op == "!=" {
			return expr + " IS NOT NULL", nil
		}
	}
	placeholder := fmt.Sprintf("$%d", argOffset+1)
	return expr + " " + op + " " + placeholder, []any{decoded}
}

func sortedKeys(m map[string]json.RawMessage) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func decodeValue(val json.RawMessage) any {
	var v any
	if err := json.Unmarshal(val, &v); err != nil {
		return nil
	}
	if b, ok := v.(bool); ok {
		if b {
			return 1
		}
		return 0
	}
	return v
}
