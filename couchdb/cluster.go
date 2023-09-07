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
	"net/http"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
)

func (c *client) ClusterStatus(ctx context.Context, options driver.Options) (string, error) {
	opts := map[string]interface{}{}
	options.Apply(opts)
	var result struct {
		State string `json:"state"`
	}
	query, err := optionsToParams(opts)
	if err != nil {
		return "", err
	}
	err = c.DoJSON(ctx, http.MethodGet, "/_cluster_setup", &chttp.Options{Query: query}, &result)
	return result.State, err
}

func (c *client) ClusterSetup(ctx context.Context, action interface{}) error {
	options := &chttp.Options{
		Body: chttp.EncodeBody(action),
	}
	_, err := c.DoError(ctx, http.MethodPost, "/_cluster_setup", options)
	return err
}

func (c *client) Membership(ctx context.Context) (*driver.ClusterMembership, error) {
	result := new(driver.ClusterMembership)
	err := c.DoJSON(ctx, http.MethodGet, "/_membership", nil, &result)
	return result, err
}
