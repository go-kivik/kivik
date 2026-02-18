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
	"testing"
)

func TestClientDBExists(t *testing.T) {
	t.Run("exists", func(t *testing.T) {
		dClient := testClient(t)
		c := dClient.(*client)

		if _, err := c.db.Exec(`CREATE TABLE "kivik$foo" (id INTEGER)`); err != nil {
			t.Fatal(err)
		}

		exists, err := dClient.DBExists(context.Background(), "foo", nil)
		if err != nil {
			t.Fatal(err)
		}
		if !exists {
			t.Fatal("foo should exist")
		}
	})
	t.Run("does not exist", func(t *testing.T) {
		dClient := testClient(t)

		exists, err := dClient.DBExists(context.Background(), "foo", nil)
		if err != nil {
			t.Fatal(err)
		}
		if exists {
			t.Fatal("foo should not exist")
		}
	})
}
