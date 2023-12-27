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

package fs

import (
	"context"
	"errors"
	"net/http"
	"os"
	"path/filepath"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/fsdb/cdb"
	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

type db struct {
	*client
	dbPath, dbName string
	fs             filesystem.Filesystem
	cdb            *cdb.FS
}

var _ driver.DB = &db{}

var notYetImplemented = statusError{status: http.StatusNotImplemented, error: errors.New("kivik: not yet implemented in fs driver")}

func (d *db) path(parts ...string) string {
	return filepath.Join(append([]string{d.dbPath}, parts...)...)
}

func (d *db) AllDocs(context.Context, driver.Options) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) Query(context.Context, string, string, driver.Options) (driver.Rows, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) CreateDoc(context.Context, interface{}, driver.Options) (docID, rev string, err error) {
	// FIXME: Unimplemented
	return "", "", notYetImplemented
}

func (d *db) Delete(context.Context, string, driver.Options) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) Stats(context.Context) (*driver.DBStats, error) {
	_, err := d.fs.Stat(d.path())
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, &statusError{status: http.StatusNotFound, error: err}
		}
		return nil, err
	}
	return &driver.DBStats{
		Name:           d.dbName,
		CompactRunning: false,
	}, nil
}

func (d *db) CompactView(context.Context, string) error {
	// FIXME: Unimplemented
	return notYetImplemented
}

func (d *db) ViewCleanup(context.Context) error {
	// FIXME: Unimplemented
	return notYetImplemented
}

func (d *db) BulkDocs(context.Context, []interface{}) ([]driver.BulkResult, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) PutAttachment(context.Context, string, *driver.Attachment, driver.Options) (string, error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) GetAttachment(context.Context, string, string, driver.Options) (*driver.Attachment, error) {
	// FIXME: Unimplemented
	return nil, notYetImplemented
}

func (d *db) DeleteAttachment(context.Context, string, string, driver.Options) (newRev string, err error) {
	// FIXME: Unimplemented
	return "", notYetImplemented
}

func (d *db) Close() error {
	return nil
}
