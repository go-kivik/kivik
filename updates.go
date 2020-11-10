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

package kivik

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
)

// DBUpdates provides access to database updates.
type DBUpdates struct {
	*iter
	updatesi driver.DBUpdates
}

// Next returns the next DBUpdate from the feed. This function will block
// until an event is received. If an error occurs, it will be returned and
// the feed closed. If the feed was closed normally, io.EOF will be returned
// when there are no more events in the buffer.
func (f *DBUpdates) Next() bool {
	return f.iter.Next()
}

// Close closes the feed. Any unread updates will still be accessible via
// Next().
func (f *DBUpdates) Close() error {
	return f.iter.Close()
}

// Err returns the error, if any, that was encountered during iteration. Err
// may be called after an explicit or implicit Close.
func (f *DBUpdates) Err() error {
	return f.iter.Err()
}

type updatesIterator struct{ driver.DBUpdates }

var _ iterator = &updatesIterator{}

func (r *updatesIterator) Next(i interface{}) error { return r.DBUpdates.Next(i.(*driver.DBUpdate)) }

func newDBUpdates(ctx context.Context, updatesi driver.DBUpdates) *DBUpdates {
	return &DBUpdates{
		iter:     newIterator(ctx, &updatesIterator{updatesi}, &driver.DBUpdate{}),
		updatesi: updatesi,
	}
}

// DBName returns the database name for the current update.
func (f *DBUpdates) DBName() string {
	runlock, err := f.rlock()
	if err != nil {
		return ""
	}
	defer runlock()
	return f.curVal.(*driver.DBUpdate).DBName
}

// Type returns the type of the current update.
func (f *DBUpdates) Type() string {
	runlock, err := f.rlock()
	if err != nil {
		return ""
	}
	defer runlock()
	return f.curVal.(*driver.DBUpdate).Type
}

// Seq returns the update sequence of the current update.
func (f *DBUpdates) Seq() string {
	runlock, err := f.rlock()
	if err != nil {
		return ""
	}
	defer runlock()
	return f.curVal.(*driver.DBUpdate).Seq
}

// DBUpdates begins polling for database updates.
func (c *Client) DBUpdates(ctx context.Context, options ...Options) (*DBUpdates, error) {
	var updaterFunc func(context.Context, map[string]interface{}) (driver.DBUpdates, error)
	switch t := c.driverClient.(type) {
	case driver.DBUpdaterWithOptions:
		updaterFunc = t.DBUpdates
	case driver.DBUpdater:
		updaterFunc = func(ctx context.Context, _ map[string]interface{}) (driver.DBUpdates, error) {
			return t.DBUpdates(ctx)
		}
	default:
		return nil, &Error{HTTPStatus: http.StatusNotImplemented, Message: "kivik: driver does not implement DBUpdater"}
	}

	updatesi, err := updaterFunc(ctx, mergeOptions(options...))
	if err != nil {
		return nil, err
	}
	return newDBUpdates(context.Background(), updatesi), nil
}
