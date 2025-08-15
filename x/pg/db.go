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

type db struct{}

var _ driver.DB = &db{}

func (d *db) AllDocs(context.Context, driver.Options) (driver.Rows, error) {
	return nil, errors.ErrUnsupported
}

func (d *db) Put(context.Context, string, interface{}, driver.Options) (string, error) {
	return "", errors.ErrUnsupported
}

func (d *db) Get(context.Context, string, driver.Options) (*driver.Document, error) {
	return nil, errors.ErrUnsupported
}

func (d *db) Delete(context.Context, string, driver.Options) (string, error) {
	return "", errors.ErrUnsupported
}

func (d *db) Stats(context.Context) (*driver.DBStats, error) {
	return nil, errors.ErrUnsupported
}

func (d *db) Compact(context.Context) error {
	return errors.ErrUnsupported
}

func (d *db) CompactView(context.Context, string) error {
	return errors.ErrUnsupported
}

func (d *db) ViewCleanup(context.Context) error {
	return errors.ErrUnsupported
}

func (d *db) Changes(context.Context, driver.Options) (driver.Changes, error) {
	return nil, errors.ErrUnsupported
}

func (d *db) PutAttachment(context.Context, string, *driver.Attachment, driver.Options) (string, error) {
	return "", errors.ErrUnsupported
}

func (d *db) GetAttachment(context.Context, string, string, driver.Options) (*driver.Attachment, error) {
	return nil, errors.ErrUnsupported
}

func (d *db) DeleteAttachment(context.Context, string, string, driver.Options) (string, error) {
	return "", errors.ErrUnsupported
}

func (d *db) Query(context.Context, string, string, driver.Options) (driver.Rows, error) {
	return nil, errors.ErrUnsupported
}

func (d *db) Close() error {
	return nil
}
