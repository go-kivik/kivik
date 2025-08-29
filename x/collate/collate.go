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

// Package collate provides (near) CouchDB-compatible collation functions.
//
// The collation order provided by this package differs slightly from that
// described by the [CouchDB documentation]. In particular:
//
//   - The Unicode UCI algorithm supported natively by Go sorts the backtick (`)
//     and caret (^) after other symbols, not before.
//   - Because Go's maps are unordered, this implementation does not honor the
//     order of object key members when collating.  That is to say, the object
//     `{b:2,a:1}` is treated as equivalent to `{a:1,b:2}` for collation
//     purposes. This is tracked in [issue #952]. Please leave a comment there
//     if this is affecting you.
//
// [CouchDB documentation]: https://docs.couchdb.org/en/stable/ddocs/views/collation.html#collation-specification
// [issue #952]: https://github.com/go-kivik/kivik/issues/952
package collate

import (
	"bytes"
	"encoding/json"
	"sort"
	"strconv"
	"sync"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

var (
	collatorMu = new(sync.Mutex)
	collator   = collate.New(language.Und)
)

// CompareString returns an integer comparing the two strings.
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
func CompareString(a, b string) int {
	collatorMu.Lock()
	defer collatorMu.Unlock()
	return collator.CompareString(a, b)
}

// CompareObject compares two unmarshaled JSON objects. The function will panic
// if it encounters an unexpected type. The comparison is performed recursively,
// with keys sorted before comparison. The result will be 0 if a==b, -1 if a < b,
// and +1 if a > b.
func CompareObject(a, b any) int {
	aType := jsonTypeOf(a)
	switch bType := jsonTypeOf(b); {
	case aType < bType:
		return -1
	case aType > bType:
		return 1
	}

	switch aType {
	case jsonTypeBool:
		aBool := a.(bool)
		bBool := b.(bool)
		if aBool == bBool {
			return 0
		}
		// false before true
		if !aBool {
			return -1
		}
		return 1
	case jsonTypeNull:
		if b == nil {
			return 0
		}
		return -1
	case jsonTypeNumber:
		return int(a.(float64) - b.(float64))
	case jsonTypeString:
		return CompareString(a.(string), b.(string))
	case jsonTypeArray:
		aArray := a.([]any)
		bArray := b.([]any)
		for i := 0; i < len(aArray) && i < len(bArray); i++ {
			if cmp := CompareObject(aArray[i], bArray[i]); cmp != 0 {
				return cmp
			}
		}
		return len(aArray) - len(bArray)
	case jsonTypeObject:
		aObject := a.(map[string]any)
		bObject := b.(map[string]any)
		keyMap := make(map[string]struct{}, len(aObject))
		for k := range aObject {
			keyMap[k] = struct{}{}
		}
		for k := range bObject {
			keyMap[k] = struct{}{}
		}
		keys := make([]string, 0, len(keyMap))
		for k := range keyMap {
			keys = append(keys, k)
		}
		sort.Slice(keys, func(i, j int) bool {
			return CompareString(keys[i], keys[j]) < 0
		})

		for i, k := range keys {
			av, aok := aObject[k]
			if !aok {
				return 1
			}
			bv, bok := bObject[k]
			if !bok {
				return -1
			}
			if cmp := CompareObject(av, bv); cmp != 0 {
				return cmp
			}
			if i+1 == len(aObject) || i+1 == len(bObject) {
				return len(aObject) - len(bObject)
			}
		}
	}
	panic("unexpected JSON type")
}

// CompareJSON compares two marshaled JSON values according to the CouchDB
// collation rules. The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
// See https://docs.couchdb.org/en/stable/ddocs/views/collation.html
func CompareJSON(a, b json.RawMessage) int {
	if bytes.Equal(a, b) {
		return 0
	}
	// Literal nothing sorts first
	if len(a) == 0 {
		return -1
	}
	if len(b) == 0 {
		return 1
	}

	at, bt := jsType(a), jsType(b)
	switch {

	// Null
	case at == jsTypeNull:
		return -1
	case bt == jsTypeNull:
		return 1

	// Booleans
	case at == jsTypeBoolean:
		if bt != jsTypeBoolean {
			return -1
		}
		if bytes.Equal(a, []byte("false")) {
			return -1
		}
		return 1
	case bt == jsTypeBoolean:
		return 1

	// Numbers
	case at == jsTypeNumber:
		if bt != jsTypeNumber {
			return -1
		}
		av, _ := strconv.ParseFloat(string(a), 64)
		bv, _ := strconv.ParseFloat(string(b), 64)
		switch {
		case av < bv:
			return -1
		case av > bv:
			return 1
		default:
			return 0
		}
	case bt == jsTypeNumber:
		return 1

	// Strings
	case at == jsTypeString:
		if bt != jsTypeString {
			return -1
		}
		return CompareString(string(a), string(b))
	case bt == jsTypeString:
		return 1

	// Arrays
	case at == jsTypeArray:
		if bt != jsTypeArray {
			return -1
		}
		var av, bv []json.RawMessage
		_ = json.Unmarshal(a, &av)
		_ = json.Unmarshal(b, &bv)
		for i := 0; i < len(av) && i < len(bv); i++ {
			if r := CompareJSON(av[i], bv[i]); r != 0 {
				return r
			}
		}
		return len(av) - len(bv)

	case bt == jsTypeArray:
		return 1

	// Objects
	case at == jsTypeObject:
		if bt != jsTypeObject {
			return -1
		}

		var av, bv rawObject
		_ = json.Unmarshal(a, &av)
		_ = json.Unmarshal(b, &bv)
		for i := 0; i < len(av) && i < len(bv); i++ {
			// First compare keys
			if r := CompareJSON(av[i][0], bv[i][0]); r != 0 {
				return r
			}
			// Then values
			if r := CompareJSON(av[i][1], bv[i][1]); r != 0 {
				return r
			}
		}

		return len(av) - len(bv)
	}

	return 1
}

// rawObject represents an ordered JSON object.
type rawObject [][2]json.RawMessage

func (r *rawObject) UnmarshalJSON(b []byte) error {
	var o map[string]json.RawMessage
	if err := json.Unmarshal(b, &o); err != nil {
		return err
	}
	*r = make([][2]json.RawMessage, 0, len(o))
	for k, v := range o {
		rawKey, _ := json.Marshal(k)
		*r = append(*r, [2]json.RawMessage{rawKey, v})
	}
	// This sort is a hack, to make sorting stable in light of the limitation
	// outlined in #952. Without this, the order is arbitrary, and the collation
	// order is unstable.  This could be simplified, but I'm leaving it as-is
	// for the moment, so that it's easy to revert to CouchDB behavior if #952
	// is ever implemented. If it is, deleting this sort call should be the
	// only change needed in the [rawObject] type.
	sort.Slice(*r, func(i, j int) bool {
		return CompareJSON((*r)[i][0], (*r)[j][0]) < 0
	})
	return nil
}

const (
	jsTypeString = iota
	jsTypeArray
	jsTypeObject
	jsTypeNull
	jsTypeBoolean
	jsTypeNumber
)

func jsType(s json.RawMessage) int {
	switch s[0] {
	case '"':
		return jsTypeString
	case '[':
		return jsTypeArray
	case '{':
		return jsTypeObject
	case 'n':
		return jsTypeNull
	case 't', 'f':
		return jsTypeBoolean
	}
	return jsTypeNumber
}

type couchdbKeys []json.RawMessage

var _ sort.Interface = &couchdbKeys{}

func (c couchdbKeys) Len() int           { return len(c) }
func (c couchdbKeys) Less(i, j int) bool { return CompareJSON(c[i], c[j]) < 0 }
func (c couchdbKeys) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
