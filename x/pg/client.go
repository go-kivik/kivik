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

	"github.com/go-kivik/kivik/v4/driver"
)

type client struct{}

var _ driver.Client = (*client)(nil)

func (c *client) Version(context.Context) (*driver.Version, error) {
	return nil, errors.ErrUnsupported
}

func (c *client) AllDBs(context.Context, driver.Options) ([]string, error) {
	return nil, errors.ErrUnsupported
}

func (c *client) DBExists(context.Context, string, driver.Options) (bool, error) {
	return false, errors.ErrUnsupported
}

func (c *client) CreateDB(context.Context, string, driver.Options) error {
	return errors.ErrUnsupported
}

func (c *client) DestroyDB(context.Context, string, driver.Options) error {
	return errors.ErrUnsupported
}

func (c *client) DB(string, driver.Options) (driver.DB, error) {
	return &db{}, nil
}
