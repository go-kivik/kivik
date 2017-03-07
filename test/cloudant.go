package test

import (
	"net/http"

	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteCloudant, kt.SuiteConfig{
		"AllDBs.expected":               []string{"_replicator", "_users"},
		"AllDBs/NoAuth.status":          http.StatusUnauthorized,
		"AllDBs/RW/group/NoAuth.status": http.StatusUnauthorized,

		"Config/Admin/GetAll.status":             http.StatusForbidden,
		"Config/Admin/GetSection.sections":       []string{"chicken"},
		"Config/Admin/GetSection/chicken.status": http.StatusForbidden,
		"Config/NoAuth/GetAll.status":            http.StatusUnauthorized,
		"Config/NoAuth/GetSection.sections":      []string{"chicken"},
		"Config/NoAuth/GetItem.items":            []string{"foo.bar"},
		"Config/NoAuth/GetSection.status":        http.StatusUnauthorized,
		"Config/NoAuth/GetItem.status":           http.StatusUnauthorized,
		"Config/RW/group/NoAuth/Set.status":      http.StatusUnauthorized,
		"Config/RW/group/Admin/Set.status":       http.StatusForbidden,
		"Config/RW/group/NoAuth/Delete.status":   http.StatusUnauthorized,
		"Config/RW/group/Admin/Delete.status":    http.StatusForbidden,

		"CreateDB/RW/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": http.StatusPreconditionFailed,

		"AllDocs/Admin.databases":            []string{"_replicator", "chicken"},
		"AllDocs/Admin/_replicator.expected": []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":   0,
		"AllDocs/Admin/chicken.status":       http.StatusNotFound,
		"AllDocs/NoAuth.databases":           []string{"_replicator", "chicken"},
		"AllDocs/NoAuth/_replicator.status":  http.StatusUnauthorized,
		"AllDocs/NoAuth/chicken.status":      http.StatusNotFound,
		"AllDocs/RW/group/NoAuth.status":     http.StatusUnauthorized,

		"DBExists.databases":              []string{"_users", "chicken"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/NoAuth/_users.status":   http.StatusUnauthorized,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.status": http.StatusUnauthorized,
	})
}
