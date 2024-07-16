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

package collate

import (
	"fmt"
	"reflect"
)

type jsonType int

const (
	jsonTypeNull jsonType = iota
	jsonTypeBool
	jsonTypeNumber
	jsonTypeString
	jsonTypeArray
	jsonTypeObject
)

func jsonTypeOf(v interface{}) jsonType {
	if isNil(v) {
		return jsonTypeNull
	}

	switch v.(type) {
	case bool:
		return jsonTypeBool
	case float64:
		return jsonTypeNumber
	case string:
		return jsonTypeString
	case []interface{}:
		return jsonTypeArray
	case map[string]interface{}:
		return jsonTypeObject
	}
	panic(fmt.Sprintf("unexpected JSON type: %T", v))
}

func isNil(v interface{}) bool {
	if v == nil {
		return true
	}
	rv := reflect.ValueOf(v)
	return rv.Kind() == reflect.Ptr && rv.IsNil()
}
