package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

func TestLog(t *testing.T) {
	s, err := kivik.New("couch", TestServerAuth)
	if err != nil {
		t.Fatalf("Error connecting to " + TestServerAuth)
	}
	logBuf := make([]byte, 0, 1000)
	_, err = s.Log(logBuf, 0)
	if err == nil {
		t.Errorf("Expected an error fetching logs from Cloudant")
	}
}

/*
func TestLog2(t *testing.T) {
	s, err := kivik.New("couch", "https://admin:xxx@argoz.flimzy.com:6984/")
	if err != nil {
		t.Fatalf("Error connecting to " + TestServerAuth)
	}
	logBuf := make([]byte, 1000)
	_, err = s.Log(logBuf, 0)
	if err != nil {
		t.Errorf("Unexpected error getting logs: %s", err)
	}
}
*/
