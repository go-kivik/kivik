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

//go:build !js

package sqlite

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

var _ driver.Cluster = (*client)(nil)

var systemDatabases = []string{"_users", "_replicator", "_global_changes"}

func (c *client) ClusterSetup(ctx context.Context, action any) error {
	var data []byte
	switch v := action.(type) {
	case json.RawMessage:
		data = v
	case []byte:
		data = v
	case string:
		data = []byte(v)
	default:
		var err error
		data, err = json.Marshal(action)
		if err != nil {
			return err
		}
	}
	var parsed struct {
		Action string `json:"action"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	switch parsed.Action {
	case "enable_single_node", "finish_cluster":
		for _, name := range systemDatabases {
			if err := c.CreateDB(ctx, name, nil); err != nil {
				if kivik.HTTPStatus(err) == http.StatusPreconditionFailed {
					continue
				}
				return err
			}
		}
		return nil
	case "enable_cluster", "add_node":
		return nil
	default:
		return fmt.Errorf("unknown action: %s", parsed.Action)
	}
}

func (c *client) ClusterStatus(context.Context, driver.Options) (string, error) {
	return "", errors.New("not implemented")
}

func (c *client) Membership(context.Context) (*driver.ClusterMembership, error) {
	return nil, errors.New("not implemented")
}
