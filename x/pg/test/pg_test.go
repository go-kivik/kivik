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

package test

import (
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
	_ "github.com/go-kivik/kivik/v4/x/pg" // Postgres driver
)

func init() {
	RegisterPGSuites()
}

func TestPG(t *testing.T) {
	t.Parallel()
	client, err := kivik.New("pg", "postgres://kivik:kivik@localhost:5432/kivik_test?sslmode=disable")
	if err != nil {
		t.Errorf("Failed to connect to PostgreSQL driver: %s", err)
		return
	}
	t.Cleanup(func() { _ = client.Close() })
	clients := &kt.Context{
		ContextCore: &kt.ContextCore{
			RW:    true,
			Admin: client,
		},
		T: t,
	}
	kiviktest.RunTestsInternal(clients, kiviktest.SuitePG)
}
