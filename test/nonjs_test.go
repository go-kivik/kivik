package test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/flimzy/kivik"
)

/*

   The FS driver tests are done here, so they don't run in GopherJS, since
   there is no FS in the browser.

*/

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
	RunSubtests(client, true, []string{SuiteKivikFS}, t)
}
