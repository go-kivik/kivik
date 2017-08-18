// +build js

package test

import (
	"testing"

	"github.com/go-kivik/pouchdb/test"
)

func TestPouchLocal(t *testing.T) {
	test.PouchLocalTest(t)
}

func TestPouchRemote(t *testing.T) {
	test.PouchRemoteTest(t)
}
