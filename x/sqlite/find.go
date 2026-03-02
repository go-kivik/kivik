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
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/x/mango"
	"github.com/go-kivik/kivik/v4/x/options"
)

const defaultFindLimit = 25

// allDocsIndex is the fallback index used when no mango index is selected.
var allDocsIndex = map[string]any{
	"ddoc": nil,
	"name": "_all_docs",
	"type": "special",
	"def": map[string]any{
		"fields": []any{map[string]any{"_id": "asc"}},
	},
}

// Explain returns the query plan for a given _find query without executing it.
func (d *db) Explain(ctx context.Context, query any, _ driver.Options) (*driver.QueryPlan, error) {
	vopts, err := options.FindOptions(query)
	if err != nil {
		return nil, err
	}

	var raw struct {
		Selector map[string]any `json:"selector"`
	}
	if err := json.Unmarshal(query.(json.RawMessage), &raw); err != nil {
		return nil, err
	}

	limit := vopts.FindLimit()
	if limit < 0 {
		limit = defaultFindLimit
	}

	var index map[string]any
	if ddoc := vopts.UseIndexDdoc(); ddoc != "" {
		index, err = d.lookupMangoIndex(ctx, ddoc, vopts.UseIndexName())
		if err != nil {
			return nil, err
		}
	}
	if index == nil {
		index, err = d.selectMangoIndex(ctx, raw.Selector, vopts.SortFields())
		if err != nil {
			return nil, err
		}
	}

	var fields []any
	if f := vopts.Fields(); len(f) > 0 {
		fields = make([]any, 0, len(f))
		for _, f := range f {
			fields = append(fields, f)
		}
	}

	bookmark := vopts.Bookmark()
	if bookmark == "" {
		bookmark = "nil"
	}

	var fieldsOpt any
	if len(fields) == 0 {
		fieldsOpt = []any{}
	} else {
		fieldsOpt = fields
	}

	var useIndex []any
	ddoc := vopts.UseIndexDdoc()
	name := vopts.UseIndexName()
	switch {
	case ddoc == "":
		useIndex = []any{}
	case name == "":
		useIndex = []any{ddoc}
	default:
		useIndex = []any{ddoc, name}
	}

	sortOpt := any(map[string]any{})
	if sf := vopts.SortFields(); len(sf) > 0 {
		sortMap := make(map[string]string, len(sf))
		for _, f := range sf {
			dir := "asc"
			if f.Desc {
				dir = "desc"
			}
			sortMap[f.Field] = dir
		}
		sortOpt = sortMap
	}

	opts := map[string]any{
		"conflicts":       vopts.Conflicts(),
		"bookmark":        bookmark,
		"sort":            sortOpt,
		"fields":          fieldsOpt,
		"limit":           limit,
		"skip":            vopts.FindSkip(),
		"r":               1,
		"update":          true,
		"stable":          false,
		"stale":           false,
		"execution_stats": false,
		"allow_fallback":  true,
		"partition":       "",
		"use_index":       useIndex,
	}

	return &driver.QueryPlan{
		DBName:   d.name,
		Selector: raw.Selector,
		Limit:    limit,
		Skip:     vopts.FindSkip(),
		Fields:   fields,
		Index:    index,
		Options:  opts,
	}, nil
}

// selectMangoIndex queries the MangoIndexes table and returns the best matching
// index whose fields cover all keys in the selector or all sort fields. When
// multiple indexes match, the one with the fewest fields is preferred. Falls
// back to allDocsIndex when no match is found.
func (d *db) selectMangoIndex(ctx context.Context, selector map[string]any, sortFields []options.SortField) (map[string]any, error) {
	rows, err := d.db.QueryContext(ctx, d.query(`SELECT ddoc, name, index_def FROM {{ .MangoIndexes }}`))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bestIndex map[string]any
	bestFieldCount := -1
	var bestDdoc string
	for rows.Next() {
		var ddoc, name, indexDef string
		if err := rows.Scan(&ddoc, &name, &indexDef); err != nil {
			return nil, err
		}
		idxFields, err := mango.ExtractIndexFields([]byte(indexDef))
		if err != nil {
			continue
		}
		if coversSelector(idxFields, selector) || coversSort(idxFields, sortFields) {
			fieldCount := len(idxFields)
			if bestFieldCount >= 0 && (fieldCount > bestFieldCount || (fieldCount == bestFieldCount && ddoc >= bestDdoc)) {
				continue
			}
			index, err := buildMangoIndexMap(ddoc, name, indexDef)
			if err != nil {
				return nil, err
			}
			bestIndex = index
			bestFieldCount = fieldCount
			bestDdoc = ddoc
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if bestIndex != nil {
		return bestIndex, nil
	}
	return allDocsIndex, nil
}

// lookupMangoIndex retrieves a specific index from the MangoIndexes table by ddoc
// and optionally by name. Returns nil if no matching index is found.
func (d *db) lookupMangoIndex(ctx context.Context, ddoc, name string) (map[string]any, error) {
	where, args := mangoIndexWhere(ddoc, name)
	q := d.query(`SELECT ddoc, name, index_def FROM {{ .MangoIndexes }} WHERE `) + where
	row := d.db.QueryRowContext(ctx, q, args...)
	var rowDdoc, rowName, indexDef string
	if err := row.Scan(&rowDdoc, &rowName, &indexDef); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return buildMangoIndexMap(rowDdoc, rowName, indexDef)
}

// buildMangoIndexMap builds a map representing a Mango index result from its
// components (ddoc, name, and index definition JSON).
func buildMangoIndexMap(ddoc, name, indexDef string) (map[string]any, error) {
	normalizedFields, err := mango.NormalizeIndexFields(indexDef)
	if err != nil {
		return nil, err
	}
	anyFields := mapFieldsToAny(normalizedFields)
	return map[string]any{
		"ddoc": ddoc,
		"name": name,
		"type": "json",
		"def":  map[string]any{"fields": anyFields},
	}, nil
}

// coversSelector reports whether the index fields cover all top-level selector keys.
func coversSelector(indexFields []string, selector map[string]any) bool {
	if len(selector) == 0 {
		return false
	}
	fieldSet := make(map[string]struct{}, len(indexFields))
	for _, f := range indexFields {
		fieldSet[f] = struct{}{}
	}
	for key := range selector {
		if strings.HasPrefix(key, "$") {
			continue
		}
		if _, ok := fieldSet[key]; !ok {
			return false
		}
	}
	return true
}

// TODO: Find should enforce a default limit of 25 when none is specified,
// matching CouchDB's _find behavior.
func (d *db) Find(ctx context.Context, query any, _ driver.Options) (driver.Rows, error) {
	vopts, err := options.FindOptions(query)
	if err != nil {
		return nil, err
	}

	var sortOrderBy string
	if sortFields := vopts.SortFields(); len(sortFields) > 0 {
		sortOrderBy, err = d.sortOrderByFromIndex(ctx, sortFields, vopts.UseIndexDdoc(), vopts.UseIndexName())
		if err != nil {
			return nil, err
		}
	}

	// TODO: CouchDB treats use_index as a hint, falling back with a warning
	// if the index doesn't exist or doesn't match the selector. This errors
	// instead. Should fall back gracefully and return a warning.
	if ddoc := vopts.UseIndexDdoc(); ddoc != "" {
		where, args := mangoIndexWhere(ddoc, vopts.UseIndexName())
		var count int
		if err := d.db.QueryRowContext(ctx, d.query(`
			SELECT COUNT(*) FROM {{ .MangoIndexes }} WHERE `)+where, args...).Scan(&count); err != nil {
			return nil, err
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

func (d *db) sortOrderByFromIndex(ctx context.Context, sortFields []options.SortField, useIndexDdoc, useIndexName string) (string, error) {
	q := d.query(`SELECT index_def FROM {{ .MangoIndexes }}`)
	var args []any
	if where, whereArgs := mangoIndexWhere(useIndexDdoc, useIndexName); where != "" {
		q += " WHERE " + where
		args = whereArgs
	}
	rows, err := d.db.QueryContext(ctx, q, args...)
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
				parts[i] = jsonExtract("view.doc", mango.FieldToJSONPath(sf.Field)) + " " + dir
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
	if len(sortFields) == 0 {
		return false
	}
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

// mapFieldsToAny converts a slice of field maps from map[string]string to map[string]any,
// suitable for inclusion in query plan definitions.
func mapFieldsToAny(fields []map[string]string) []any {
	anyFields := make([]any, len(fields))
	for i, f := range fields {
		m := make(map[string]any, len(f))
		for k, v := range f {
			m[k] = v
		}
		anyFields[i] = m
	}
	return anyFields
}

// mangoIndexWhere builds a WHERE clause and args for querying the mango
// indexes table, filtering by ddoc and optionally by name.
func mangoIndexWhere(ddoc, name string) (string, []any) {
	if ddoc == "" {
		return "", nil
	}
	if name != "" {
		return "ddoc = $1 AND name = $2", []any{ddoc, name}
	}
	return "ddoc = $1", []any{ddoc}
}

// selectorToSQL translates a Mango selector JSON object into parameterized SQL
// WHERE conditions. It returns condition strings using json_extract(doc.doc, ...)
// expressions and corresponding bind values. argOffset sets the starting $N
// placeholder number so conditions can be appended to an existing argument list.
// Unsupported operators are silently skipped, broadening the result set for the
// in-memory selector.Match() safety net to correct.
func selectorToSQL(selector json.RawMessage, argOffset int) ([]string, []any, bool) {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(selector, &fields); err != nil {
		return nil, nil, true
	}

	var conds []string
	var args []any
	complete := true

	for _, key := range sortedKeys(fields) {
		val := fields[key]
		switch {
		case key == "$and":
			c, a, ok := combineSelectors(val, " AND ", false, argOffset+len(args))
			conds = append(conds, c...)
			args = append(args, a...)
			if !ok {
				complete = false
			}

		case key == "$or":
			c, a, ok := combineSelectors(val, " OR ", true, argOffset+len(args))
			conds = append(conds, c...)
			args = append(args, a...)
			if !ok {
				complete = false
			}

		case strings.HasPrefix(key, "$"):
			continue

		default:
			jsonPath := mango.FieldToJSONPath(key)
			c, a := fieldCondition(jsonPath, val, argOffset+len(args))
			if c != "" {
				conds = append(conds, c)
				args = append(args, a...)
			} else if len(val) > 0 && val[0] == '{' {
				complete = false
			}
		}
	}

	if len(conds) == 0 {
		return nil, nil, complete
	}
	return conds, args, complete
}

// combineSelectors unmarshals val as an array of sub-selectors, converts each
// to SQL, and joins them with sep. If wrap is true, the result is parenthesized.
func combineSelectors(val json.RawMessage, sep string, wrap bool, argOffset int) ([]string, []any, bool) {
	var elements []json.RawMessage
	if err := json.Unmarshal(val, &elements); err != nil {
		return nil, nil, true
	}
	var parts []string
	var args []any
	complete := true
	for _, elem := range elements {
		subConds, subArgs, ok := selectorToSQL(elem, argOffset+len(args))
		parts = append(parts, subConds...)
		args = append(args, subArgs...)
		if !ok {
			complete = false
		}
	}
	if len(parts) == 0 {
		return nil, nil, complete
	}
	joined := strings.Join(parts, sep)
	if wrap {
		joined = "(" + joined + ")"
	}
	return []string{joined}, args, complete
}

func jsonExtract(col, jsonPath string) string {
	return `json_extract(` + col + `, '` + strings.ReplaceAll(jsonPath, "'", "''") + `')`
}

func jsonType(col, jsonPath string) string {
	return `json_type(` + col + `, '` + strings.ReplaceAll(jsonPath, "'", "''") + `')`
}

func fieldCondition(jsonPath string, val json.RawMessage, argOffset int) (string, []any) {
	if len(val) == 0 {
		return "", nil
	}
	expr := jsonExtract("doc.doc", jsonPath)
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
	typeExpr := jsonType("doc.doc", jsonPath)

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
