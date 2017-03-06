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

		"Config/Admin/GetAll.status":             http.StatusForbidden,
		"Config/Admin/GetSection.sections":       []string{"chicken"},
		"Config/Admin/GetSection/chicken.status": http.StatusForbidden,
		"Config/NoAuth/GetAll.status":            http.StatusUnauthorized,
		"Config/NoAuth/GetSection.sections":      []string{"chicken"},
		"Config/NoAuth/GetItem.items":            []string{"foo.bar"},
		"Config/NoAuth/GetSection.status":        http.StatusUnauthorized,
		"Config/NoAuth/GetItem.status":           http.StatusUnauthorized,
		"Config/RW/NoAuth/Set.status":            http.StatusUnauthorized,
		"Config/RW/Admin/Set.status":             http.StatusForbidden,
		"Config/RW/NoAuth/Delete.status":         http.StatusUnauthorized,
		"Config/RW/Admin/Delete.status":          http.StatusForbidden,

		"CreateDB/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/Admin/Recreate.status": http.StatusPreconditionFailed,
	})
}
