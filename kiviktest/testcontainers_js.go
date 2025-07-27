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

//go:build js

package kiviktest

import (
	"bytes"
	"io"
	"net/http"
	"testing"
)

func startCouchDB(t *testing.T, image string) string { //nolint:thelper // Not a helper
	t.Logf("testcontainers: Starting CouchDB with image: %s", image)
	req, err := http.NewRequest(http.MethodPost, "http://localhost:8080?image="+image, nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("Expected status OK, got %s", resp.Status)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	dsn := string(bytes.TrimSpace(body))
	if dsn == "" {
		t.Fatal("Received empty DSN from CouchDB daemon")
	}
	t.Logf("testcontainers: CouchDB started with DSN: %s", dsn)
	return dsn
}
