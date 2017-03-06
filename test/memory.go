package test

import (
	"net/http"

	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteKivikMemory, kt.SuiteConfig{
		"AllDBs.expected": []string{},
		"Config.status":   http.StatusNotImplemented,
	})
}
