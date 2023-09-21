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

package kivik

import (
	"encoding/json"
)

// document represents any CouchDB document.
type document struct {
	ID          string                 `json:"_id"`
	Rev         string                 `json:"_rev"`
	Attachments *Attachments           `json:"_attachments,omitempty"`
	Data        map[string]interface{} `json:"-"`
}

// MarshalJSON satisfies the json.Marshaler interface
func (d *document) MarshalJSON() ([]byte, error) {
	var data []byte
	doc, err := json.Marshal(*d)
	if err != nil {
		return nil, err
	}
	if len(d.Data) > 0 {
		var err error
		if data, err = json.Marshal(d.Data); err != nil {
			return nil, err
		}
		doc[len(doc)-1] = ','
		return append(doc, data[1:]...), nil
	}
	return doc, nil
}

// UnmarshalJSON satisfies the json.Unmarshaler interface.
func (d *document) UnmarshalJSON(p []byte) error {
	type internalDoc document
	doc := &internalDoc{}
	if err := json.Unmarshal(p, &doc); err != nil {
		return err
	}
	data := make(map[string]interface{})
	if err := json.Unmarshal(p, &data); err != nil {
		return err
	}
	delete(data, "_id")
	delete(data, "_rev")
	delete(data, "_attachments")
	*d = document(*doc)
	d.Data = data
	return nil
}
