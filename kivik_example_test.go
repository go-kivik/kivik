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

package kivik_test

import (
	"fmt"

	kivik "github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/couchdb" // CouchDB driver, needed for executable example
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func init() {
	kivik.Register("driver", &mock.Driver{
		NewClientFunc: func(_ string, _ driver.Options) (driver.Client, error) {
			return &mock.Client{
				DBFunc: func(string, driver.Options) (driver.DB, error) {
					return nil, nil
				},
			}, nil
		},
	})
}

// New is used to create a client handle. `driver` specifies the name of the
// registered database driver and `dataSourceName` specifies the
// database-specific connection information, such as a URL.
func ExampleNew() {
	client, err := kivik.New("driver", "http://example.com:5984/")
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to", client.DSN())
	// Output: Connected to http://example.com:5984/
}

// With a client handle in hand, you can create a database handle with the DB()
// method to interact with a specific database.
func Example_connecting() {
	client, err := kivik.New("couch", "http://example.com:5984/")
	if err != nil {
		panic(err)
	}
	db := client.DB("_users")
	fmt.Println("Database handle for " + db.Name())
	// Output: Database handle for _users
}
