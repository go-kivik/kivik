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
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func startCouchDB(t *testing.T, image string) string { //nolint:thelper // Not a helper
	if os.Getenv("USETC") == "" {
		t.Skip("USETC not set, skipping testcontainers")
	}
	addr := startTCDaemon(t)
	t.Logf("testcontainers: Starting CouchDB with image: %s", image)
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s?image=%s", addr, image), nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("Expected status OK, got %s. Response body: %s", resp.Status, string(body))
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

func startTCDaemon(t *testing.T) string {
	t.Helper()
	t.Log("Starting testcontainers daemon...")

	cmd := exec.Command("go", "run", "./cmd")
	cmd.Dir = tcModuleDir(t)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to get stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to get stderr pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("Failed to start testcontainers daemon: %v", err)
	}

	t.Cleanup(func() {
		if cmd.Process != nil {
			_ = cmd.Process.Kill()
		}
	})

	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			t.Logf("[STDERR] %s", scanner.Text())
		}
	}()

	ready := make(chan string, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			t.Logf("[STDOUT] %s", line)
			if addr, ok := strings.CutPrefix(line, "Listening on "); ok {
				ready <- addr
				close(ready)
				return
			}
		}
	}()

	addr, ok := <-ready
	if !ok {
		t.Fatal("testcontainers daemon exited without providing address")
	}
	return addr
}
