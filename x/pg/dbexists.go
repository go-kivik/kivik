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

package pg

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func (c *client) DBExists(ctx context.Context, dbName string, _ driver.Options) (bool, error) {
	var exists bool
	err := c.pool.QueryRow(ctx, "SELECT to_regclass($1) IS NOT NULL", dbName).Scan(&exists)
	if err != nil {
		return false, err
	}
	if exists {
		return true, nil
	}
	return false, &internal.Error{
		Status:  http.StatusNotFound,
		Message: fmt.Sprintf("database %q not found", dbName),
	}
}
