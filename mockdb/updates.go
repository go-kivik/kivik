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

package mockdb

import (
	"context"
	"time"

	"github.com/go-kivik/kivik/v4/driver"
)

// Updates is a mocked collection of database updates.
type Updates struct {
	iter
	lastSeq    string
	lastSeqErr error
}

type driverDBUpdates struct {
	context.Context
	*Updates
}

func coalesceDBUpdates(updates *Updates) *Updates {
	if updates != nil {
		return updates
	}
	return &Updates{}
}

var _ driver.DBUpdates = &driverDBUpdates{}

func (u *driverDBUpdates) Next(update *driver.DBUpdate) error {
	result, err := u.unshift(u.Context)
	if err != nil {
		return err
	}
	*update = *result.(*driver.DBUpdate)
	return nil
}

func (u *driverDBUpdates) LastSeq() (string, error) {
	return u.lastSeq, u.lastSeqErr
}

// CloseError sets an error to be returned when the updates iterator is closed.
func (u *Updates) CloseError(err error) *Updates {
	u.closeErr = err
	return u
}

// AddUpdateError adds an error to be returned during update iteration.
func (u *Updates) AddUpdateError(err error) *Updates {
	u.resultErr = err
	return u
}

// AddUpdate adds a database update to be returned by the DBUpdates iterator. If
// AddUpdateError has been set, this method will panic.
func (u *Updates) AddUpdate(update *driver.DBUpdate) *Updates {
	if u.resultErr != nil {
		panic("It is invalid to set more updates after AddUpdateError is defined.")
	}
	u.push(&item{item: update})
	return u
}

// AddDelay adds a delay before the next iteration will complete.
func (u *Updates) AddDelay(delay time.Duration) *Updates {
	u.push(&item{delay: delay})
	return u
}

// LastSeq sets the LastSeq value to be returned by the DBUpdates iterator.
func (u *Updates) LastSeq(lastSeq string) *Updates {
	u.lastSeq = lastSeq
	return u
}

// LastSeqError sets the error value to be returned when LastSeq is called.
func (u *Updates) LastSeqError(err error) *Updates {
	u.lastSeqErr = err
	return u
}

// Final converts the Updates object to a driver.DBUpdates. This method is
// intended for use within WillExecute() to return results.
func (u *Updates) Final() driver.DBUpdates {
	return &driverDBUpdates{Updates: u}
}
