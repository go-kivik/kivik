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

package pg

import (
	"context"
	"sync"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testUser     = "kivik"
	testPassword = "kivik"
)

var startPostgresOnce = sync.OnceValues(startPostgresContainer)

// startPostgres starts PostgreSQL in a Docker container, and returns
// the DSN.
func startPostgres() (string, error) {
	return startPostgresOnce()
}

func startPostgresContainer() (string, error) {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx, "postgres:17.6-alpine3.21",
		postgres.WithUsername(testUser),
		postgres.WithPassword(testPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second)),
	)
	if err != nil {
		return "", err
	}

	return postgresContainer.ConnectionString(ctx)
}
