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
	"context"

	"github.com/go-kivik/kivik/v4/driver"
)

type drv struct{}

var _ driver.Driver = (*drv)(nil)

func (drv) NewClient(name string, _ driver.Options) (driver.Client, error) {
	return &client{}, nil
}

type client struct{}

var _ driver.Client = (*client)(nil)

func (client) Version(context.Context) (*driver.Version, error) {
	return nil, nil
}

func (client) AllDBs(context.Context, driver.Options) ([]string, error) {
	return nil, nil
}

func (client) DBExists(context.Context, string, driver.Options) (bool, error) {
	return false, nil
}

func (client) CreateDB(context.Context, string, driver.Options) error {
	return nil
}

func (client) DestroyDB(context.Context, string, driver.Options) error {
	return nil
}

func (client) DB(string, driver.Options) (driver.DB, error) {
	return nil, nil
}
