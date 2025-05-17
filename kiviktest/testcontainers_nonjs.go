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

//go:build !js

package kiviktest

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func startCouchDB(t *testing.T, image string) string { //nolint:thelper // Not a helper
	if os.Getenv("USETC") == "" {
		t.Skip("USETC not set, skipping testcontainers")
	}
	req := testcontainers.ContainerRequest{
		Image:        image,
		ExposedPorts: []string{"5984/tcp"},
		WaitingFor:   wait.ForHTTP("/").WithPort("5984/tcp").WithStartupTimeout(120 * time.Second),
		Env: map[string]string{
			"COUCHDB_USER":     "admin",
			"COUCHDB_PASSWORD": "abc123",
		},
	}
	container, err := testcontainers.GenericContainer(context.TODO(), testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatal(err)
	}
	ip, err := container.Host(context.TODO())
	if err != nil {
		t.Fatal(err)
	}
	mappedPort, err := container.MappedPort(context.TODO(), "5984/tcp")
	if err != nil {
		t.Fatal(err)
	}
	dsn := fmt.Sprintf("http://admin:abc123@%s:%s", ip, mappedPort.Port())
	for _, db := range []string{"_replicator", "_users", "_global_changes"} {
		put(t, dsn+"/"+db, nil)
	}
	put(t, dsn+"/_node/nonode@nohost/_config/replicator/interval", strings.NewReader(`"1000"`))
	put(t, dsn+"/_node/nonode@nohost/_config/replicator/worker_processes", strings.NewReader(`"1"`))
	return dsn
}

func put(t *testing.T, path string, body io.Reader) {
	t.Helper()
	rq, err := http.NewRequest(http.MethodPut, path, body)
	if err != nil {
		t.Fatal(err)
	}
	rq.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(rq)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated, http.StatusPreconditionFailed:
		return
	}
	t.Fatalf("Failed to create %s: %s", path, resp.Status)
}
