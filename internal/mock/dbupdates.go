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

package mock

import "github.com/go-kivik/kivik/v4/driver"

// DBUpdates mocks driver.DBUpdates
type DBUpdates struct {
	// ID identifies a specific DBUpdates instance.
	ID        string
	NextFunc  func(*driver.DBUpdate) error
	CloseFunc func() error
}

var _ driver.DBUpdates = &DBUpdates{}

// Next calls u.NextFunc
func (u *DBUpdates) Next(dbupdate *driver.DBUpdate) error {
	return u.NextFunc(dbupdate)
}

// Close calls u.CloseFunc
func (u *DBUpdates) Close() error {
	if u.CloseFunc != nil {
		return u.CloseFunc()
	}
	return nil
}
