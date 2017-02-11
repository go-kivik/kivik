package test

// FIXME: This test disabled until a bug in the alldbs plugin can be fixed, or
// a workaround discovered. See https://github.com/nolanlawson/pouchdb-all-dbs/issues/25

// import (
// 	"reflect"
// 	"testing"
//
// 	"github.com/flimzy/kivik"
// )
//
// var ExpectedAllDBs []string
//
// func TestAllDBs(t *testing.T) {
// 	s, err := kivik.New("memdown", TestServer)
// 	if err != nil {
// 		t.Fatalf("Error connecting to %s: %s\n", TestServer, err)
// 	}
// 	allDBs, err := s.AllDBs()
// 	if err != nil {
// 		t.Fatalf("Failed to get all DBs: %s", err)
// 	}
// 	if !reflect.DeepEqual(ExpectedAllDBs, allDBs) {
// 		t.Errorf("All DBs.\n\tExpected: %v\n\t  Actual: %v\n", ExpectedAllDBs, allDBs)
// 	}
// }
