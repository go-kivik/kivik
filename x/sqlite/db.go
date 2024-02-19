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
	"database/sql"

	"github.com/go-kivik/kivik/v4/driver"
)

type db struct {
	db   *sql.DB
	name string
}

var _ driver.DB = (*db)(nil)

func (db) CreateDoc(context.Context, interface{}, driver.Options) (string, string, error) {
	return "", "", nil
}

func (db) Stats(context.Context) (*driver.DBStats, error) {
	return nil, nil
}

func (db) Compact(context.Context) error {
	return nil
}

func (db) CompactView(context.Context, string) error {
	return nil
}

func (db) ViewCleanup(context.Context) error {
	return nil
}

func (db) Changes(context.Context, driver.Options) (driver.Changes, error) {
	return nil, nil
}

func (db) PutAttachment(context.Context, string, *driver.Attachment, driver.Options) (string, error) {
	return "", nil
}

func (db) GetAttachment(context.Context, string, string, driver.Options) (*driver.Attachment, error) {
	return nil, nil
}

func (db) DeleteAttachment(context.Context, string, string, driver.Options) (string, error) {
	return "", nil
}

func (db) Query(context.Context, string, string, driver.Options) (driver.Rows, error) {
	return nil, nil
}

func (db) Close() error {
	return nil
}
