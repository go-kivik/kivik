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
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
)

// Version returns the server's version info.
func (c *client) Version(ctx context.Context) (*driver.Version, error) {
	i := &info{}
	err := c.DoJSON(ctx, http.MethodGet, "/", nil, i)
	return &driver.Version{
		Version:     i.Version,
		Vendor:      i.Vendor.Name,
		Features:    i.Features,
		RawResponse: i.Data,
	}, err
}

type info struct {
	Data     json.RawMessage
	Version  string   `json:"version"`
	Features []string `json:"features"`
	Vendor   struct {
		Name string `json:"name"`
	} `json:"vendor"`
}

func (i *info) UnmarshalJSON(data []byte) error {
	type alias info
	var a alias
	if err := json.Unmarshal(data, &a); err != nil {
		return err
	}
	i.Data = data
	i.Version = a.Version
	i.Vendor = a.Vendor
	i.Features = a.Features
	return nil
}
