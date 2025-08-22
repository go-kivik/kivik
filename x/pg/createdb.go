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
	"regexp"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

var validDBNameRE = regexp.MustCompile("^[a-z_][a-z0-9_$()+/-]*$")

func (c *client) CreateDB(ctx context.Context, dbName string, _ driver.Options) error {
	if !validDBNameRE.MatchString(dbName) {
		return &internal.Error{
			Status:  http.StatusBadRequest,
			Message: fmt.Sprintf("invalid database name: %q", dbName),
		}
	}

	_, err := c.pool.Exec(ctx, "CREATE TABLE "+dbName+" (id SERIAL PRIMARY KEY, data JSONB)")
	return err
}
