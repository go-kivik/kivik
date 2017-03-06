package test

import (
	"net/http"

	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteCloudant, kt.SuiteConfig{
		"AllDBs.expected":         []string{"_replicator", "_users"},
		"AllDBs/NoAuth.status":    http.StatusUnauthorized,
		"AllDBs/RW/NoAuth.status": http.StatusUnauthorized,
	})
}
