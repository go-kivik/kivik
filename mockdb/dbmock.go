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

import "sync"

// DB serves to create expectations for database actions to mock and test real
// database behavior.
type DB struct {
	name   string
	id     int
	client *Client
	count  int
	mu     sync.RWMutex
}

func (db *DB) expectations() int {
	return db.count
}

// ExpectClose queues an expectation for DB.Close() to be called.
func (db *DB) ExpectClose() *ExpectedDBClose {
	e := &ExpectedDBClose{
		commonExpectation: commonExpectation{db: db},
	}
	db.count++
	db.client.expected = append(db.client.expected, e)
	return e
}
