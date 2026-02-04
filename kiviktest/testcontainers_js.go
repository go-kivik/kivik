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
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"syscall/js"
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

// startTCDaemon starts the testcontainers daemon in a separate
// process, and returns the listening address.
func startTCDaemon(t *testing.T) string {
	return <-spawnTCDaemon(t)
}

func spawnTCDaemon(t *testing.T) <-chan string {
	t.Helper()

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("Failed to start testcontainers daemon: %v", r)
		}
	}()

	t.Log("Starting testcontainers daemon...")
	spawn := js.Global().Call("require", "child_process").Get("spawn")

	options := js.Global().Get("Object").New()
	options.Set("detached", true)
	stdio := js.Global().Get("Array").New(3)
	stdio.SetIndex(0, "ignore")
	stdio.SetIndex(1, "pipe")
	stdio.SetIndex(2, "pipe")
	options.Set("stdio", stdio)
	options.Set("env", js.Global().Get("process").Get("env"))
	options.Set("cwd", js.Global().Get("process").Call("cwd"))

	args := js.Global().Get("Array").New(2)
	args.SetIndex(0, "run")
	args.SetIndex(1, "github.com/go-kivik/kivik/v4/kiviktest/testcontainers/cmd")

	child := spawn.Invoke("go", args, options)

	// Add error event listener for spawn failures
	child.Call("on", "error", js.FuncOf(func(this js.Value, args []js.Value) any {
		err := args[0]
		t.Fatalf("Child process error: %s", err.String())
		return nil
	}))

	// Add exit event listener
	child.Call("on", "exit", js.FuncOf(func(this js.Value, args []js.Value) any {
		code := args[0]
		if !code.IsNull() && code.Int() != 0 {
			t.Fatalf("Child process exited with code: %d", code.Int())
		}
		return nil
	}))

	ready := make(chan string)

	// Log stdout
	child.Get("stdout").Call("on", "data", js.FuncOf(func(this js.Value, args []js.Value) any {
		buffer := args[0]
		data := buffer.Call("toString").String()
		go t.Logf("[STDOUT] %s", data)
		if strings.HasPrefix(data, "Listening on ") {
			// Extract the address from the log message
			addr := strings.TrimSpace(strings.TrimPrefix(data, "Listening on "))
			go func() {
				ready <- addr
				close(ready)
			}()
		}
		return nil
	}))

	// Log stderr
	child.Get("stderr").Call("on", "data", js.FuncOf(func(this js.Value, args []js.Value) any {
		buffer := args[0]
		data := buffer.Call("toString").String()
		go t.Logf("[STDERR] %s", data)
		return nil
	}))

	child.Call("unref") // Let the child keep running independently of the parent
	t.Cleanup(func() {
		js.Global().Get("process").Call("kill", child.Get("pid"))
	})
	return ready
}
