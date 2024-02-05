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
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-kivik/kivik/v4/internal"
)

// marshalDoc marshals doc without the _id and _rev fields, returning those,
// if they exist, as separate return values.
//
// TODO: Optimize this function.
func marshalDoc(doc interface{}) (_ []byte, id, revID, rev string, _ error) {
	tmpJSON, err := json.Marshal(doc)
	if err != nil {
		return nil, "", "", "", &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	var meta struct {
		ID  string `json:"_id"`
		Rev string `json:"_rev"`
	}
	if err := json.Unmarshal(tmpJSON, &meta); err != nil {
		return nil, "", "", "", &internal.Error{Status: http.StatusBadRequest, Err: err}
	}
	var tmp map[string]interface{}
	_ = json.Unmarshal(tmpJSON, &tmp)
	delete(tmp, "_id")
	delete(tmp, "_rev")
	finalJSON, _ := json.Marshal(tmp)
	if meta.Rev != "" {
		revParts := strings.SplitN(meta.Rev, "-", 2)
		return finalJSON, meta.ID, revParts[0], revParts[1], nil
	}
	md5 := md5.Sum(finalJSON)
	rev = hex.EncodeToString(md5[:])
	return finalJSON, meta.ID, "", rev, nil
}
