// +build !js

package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/flimzy/kivik"
	_ "github.com/flimzy/kivik/driver/fs"
	"github.com/flimzy/kivik/test/kt"
)

func TestFS(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "kivik.test.")
	if err != nil {
		t.Errorf("Failed to create temp dir to test FS driver: %s\n", err)
		return
	}
	os.RemoveAll(tempDir)       // So the driver can re-create it as desired
	defer os.RemoveAll(tempDir) // To clean up after tests
	client, err := kivik.New("fs", tempDir)
	if err != nil {
		t.Errorf("Failed to connect to FS driver: %s\n", err)
		return
	}
	clients := &kt.Context{
		Admin: client,
	}
	runTests(clients, SuiteKivikFS, t)
}
