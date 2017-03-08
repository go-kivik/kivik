// +build !js

package test

import (
	"testing"

	_ "github.com/flimzy/kivik/driver/couchdb"
	_ "github.com/flimzy/kivik/driver/fs"
	_ "github.com/flimzy/kivik/driver/memory"
)

func TestCouch16(t *testing.T) {
	doTest(SuiteCouch16, "KIVIK_TEST_DSN_COUCH16", t)
}

func TestCloudant(t *testing.T) {
	doTest(SuiteCloudant, "KIVIK_TEST_DSN_CLOUDANT", t)
}

func TestCouch20(t *testing.T) {
	doTest(SuiteCouch20, "KIVIK_TEST_DSN_COUCH20", t)
}
