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
	"net"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Postgres driver for pgtestdb
	"github.com/peterldowns/pgtestdb"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	testDatabase = "kivik"
	testUser     = "kivik"
	testPassword = "kivik"
)

func newPostgres(t *testing.T) string {
	t.Helper()
	if os.Getenv("USETC") == "" {
		t.Skip("USETC not set, skipping testcontainers")
	}
	dsn, err := startPostgres()
	if err != nil {
		t.Fatalf("Failed to start PostgreSQL container: %s", err)
	}

	parsedDSN, err := url.Parse(dsn)
	if err != nil {
		t.Fatal(err)
	}
	host, portStr, err := net.SplitHostPort(parsedDSN.Host)
	if err != nil {
		t.Fatal(err)
	}

	dbconf := pgtestdb.Config{
		DriverName: "pgx",
		User:       testUser,
		Password:   testPassword,
		Database:   testDatabase,
		Host:       host,
		Port:       portStr,
		Options:    "sslmode=disable",
		TestRole: &pgtestdb.Role{
			Username:     "postgres",
			Password:     "postgres",
			Capabilities: "SUPERUSER",
		},
	}
	config := pgtestdb.Custom(t, dbconf, pgtestdb.NoopMigrator{})

	return config.URL()
}

var startPostgresOnce = sync.OnceValues(startPostgresContainer)

// startPostgres starts PostgreSQL in a Docker container, and returns
// the DSN.
func startPostgres() (string, error) {
	return startPostgresOnce()
}

func startPostgresContainer() (string, error) {
	ctx := context.Background()
	postgresContainer, err := postgres.Run(ctx, "postgres:17.6-alpine3.21",
		postgres.WithDatabase(testDatabase),
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
