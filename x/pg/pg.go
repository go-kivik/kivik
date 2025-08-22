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
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

type pg struct{}

var _ driver.Driver = &pg{}

func init() {
	kivik.Register("pg", &pg{})
}

func (*pg) NewClient(dsn string, _ driver.Options) (driver.Client, error) {
	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, &internal.Error{
			Status: http.StatusBadRequest,
			Err:    err,
		}
	}
	if config.ConnConfig.Database == "" {
		return nil, &internal.Error{
			Status: http.StatusBadRequest,
			Err:    errors.New("database name is required in DSN"),
		}
	}
	pool, err := pgxpool.NewWithConfig(context.TODO(), config)
	if err != nil {
		return nil, &internal.Error{
			Status: http.StatusInternalServerError,
			Err:    err,
		}
	}
	return &client{pool: pool}, nil
}
