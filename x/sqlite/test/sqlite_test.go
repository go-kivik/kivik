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
	"os"
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
	_ "github.com/go-kivik/kivik/x/sqlite/v4" // SQLite driver
)

func init() {
	RegisterSQLiteSuites()
}

func TestSQLite(t *testing.T) {
	t.Parallel()
	f, err := os.CreateTemp("", "kivik-sqlite-test-*.db")
	if err != nil {
		t.Fatal(err)
	}
	dsn := f.Name() + "?_txlock=immediate&_pragma=busy_timeout(5000)&_pragma=journal_mode(WAL)"
	_ = f.Close()
	t.Cleanup(func() { _ = os.Remove(dsn) })
	client, err := kivik.New("sqlite", dsn)
	if err != nil {
		t.Errorf("Failed to connect to SQLite driver: %s", err)
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
	kiviktest.RunTestsInternal(clients, kiviktest.SuiteKivikSQLite)
}
