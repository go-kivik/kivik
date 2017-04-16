package fs

import (
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
	"github.com/flimzy/kivik/test/kt"
)

func TestCreateDBUnauthorized(t *testing.T) {
	path := "/this/better/not/exist"
	_, err := kivik.New(kt.CTX, "fs", path)
	if err == nil {
		t.Errorf("Expected error attempting to create FS database in '%s'\n", path)
		return
	}
	if errors.StatusCode(err) != kivik.StatusUnauthorized {
		t.Errorf("Expected Unauthorized error trying to create FS database in '%s', but got %s\n", path, err)
	}
}
