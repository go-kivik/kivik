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

// Package registry handles driver registrations. It's in a separate package
// to facilitate testing.
package registry

import (
	"sync"

	"github.com/go-kivik/kivik/v4/driver"
)

var (
	driversMu sync.RWMutex
	drivers   = make(map[string]driver.Driver)
)

// Register makes a database driver available by the provided name. If Register
// is called twice with the same name or if driver is nil, it panics.
//
// driver must be a driver.Driver or driver.LegacyDriver.
func Register(name string, drv driver.Driver) {
	if drv == nil {
		panic("kivik: Register driver is nil")
	}
	driversMu.Lock()
	defer driversMu.Unlock()
	if _, dup := drivers[name]; dup {
		panic("kivik: Register called twice for driver " + name)
	}
	drivers[name] = drv
}

// Driver returns the driver registered with the requested name, or nil if
// it has not been registered.
func Driver(name string) driver.Driver {
	driversMu.RLock()
	defer driversMu.RUnlock()
	return drivers[name]
}
