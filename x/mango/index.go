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

package mango

import (
	"encoding/json"
	"strings"
)

// FieldToJSONPath converts a Mango field name to a JSON path expression using
// dot-quoted notation (e.g. `$."name"`, `$."address"."city"`). Double quotes
// and backslashes within segments are escaped with a backslash.
func FieldToJSONPath(field string) string {
	segments := SplitKeys(field)
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`)
	var b strings.Builder
	b.WriteByte('$')
	for _, seg := range segments {
		b.WriteString(`."`)
		_, _ = replacer.WriteString(&b, seg)
		b.WriteByte('"')
	}
	return b.String()
}

// ExtractIndexFields parses a Mango index definition and returns the ordered
// field names. Direction is ignored since SQLite can traverse B-tree indexes in
// either direction.
func ExtractIndexFields(indexDef []byte) ([]string, error) {
	var def struct {
		Fields []any `json:"fields"`
	}
	if err := json.Unmarshal(indexDef, &def); err != nil {
		return nil, err
	}

	fields := make([]string, 0, len(def.Fields))
	for _, f := range def.Fields {
		switch v := f.(type) {
		case string:
			fields = append(fields, v)
		case map[string]any:
			for k := range v {
				fields = append(fields, k)
			}
		}
	}

	return fields, nil
}

// NormalizeIndexFields parses a stored JSON index definition and expands
// shorthand field names (e.g. "name") into the explicit {"name": "asc"} form
// returned by GetIndexes.
func NormalizeIndexFields(indexDef string) ([]map[string]string, error) {
	var def struct {
		Fields []any `json:"fields"`
	}
	if err := json.Unmarshal([]byte(indexDef), &def); err != nil {
		return nil, err
	}

	normalized := make([]map[string]string, 0, len(def.Fields))
	for _, f := range def.Fields {
		switch v := f.(type) {
		case string:
			normalized = append(normalized, map[string]string{v: "asc"})
		case map[string]any:
			m := make(map[string]string)
			for k, val := range v {
				m[k] = val.(string)
			}
			normalized = append(normalized, m)
		}
	}

	return normalized, nil
}
