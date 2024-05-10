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
	"bytes"
	"encoding/json"
	"slices"
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

func couchdbCmpString(a, b string) int {
	return couchdbCmpJSON(json.RawMessage(a), json.RawMessage(b))
}

// couchdbCmpJSON is a comparison function for CouchDB collation.
// See https://docs.couchdb.org/en/stable/ddocs/views/collation.html
func couchdbCmpJSON(a, b json.RawMessage) int {
	if bytes.Equal(a, b) {
		return 0
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
		collatorMu.Lock()
		cmp := collator.CompareString(string(a), string(b))
		collatorMu.Unlock()
		return cmp
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
			if r := couchdbCmpJSON(av[i], bv[i]); r != 0 {
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
			if r := couchdbCmpJSON(av[i][0], bv[i][0]); r != 0 {
				return r
			}
			// Then values
			if r := couchdbCmpJSON(av[i][1], bv[i][1]); r != 0 {
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
	slices.SortFunc(*r, func(a, b [2]json.RawMessage) int {
		if r := couchdbCmpJSON(a[0], b[0]); r != 0 {
			return r
		}
		return couchdbCmpJSON(a[1], b[1])
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
func (c couchdbKeys) Less(i, j int) bool { return couchdbCmpJSON(c[i], c[j]) < 0 }
func (c couchdbKeys) Swap(i, j int)      { c[i], c[j] = c[j], c[i] }
