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

package couchdb

import (
	"encoding/json"
	"net/http"

	internal "github.com/go-kivik/kivik/v4/int/errors"
)

// encodeKey encodes a key to a view query, or similar, to be passed to CouchDB.
func encodeKey(i interface{}) (string, error) {
	if raw, ok := i.(json.RawMessage); ok {
		return string(raw), nil
	}
	raw, err := json.Marshal(i)
	if err != nil {
		err = &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	return string(raw), err
}

var jsonKeys = []string{"endkey", "end_key", "key", "startkey", "start_key", "keys", "doc_ids"}

func encodeKeys(opts map[string]interface{}) error {
	for _, key := range jsonKeys {
		if v, ok := opts[key]; ok {
			value, err := encodeKey(v)
			if err != nil {
				return err
			}
			opts[key] = value
		}
	}
	return nil
}
