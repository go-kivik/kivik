package test

import (
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteCloudant, kt.SuiteConfig{
		"AllDBs.expected":               []string{"_replicator", "_users"},
		"AllDBs/NoAuth.status":          kivik.StatusUnauthorized,
		"AllDBs/RW/group/NoAuth.status": kivik.StatusUnauthorized,

		"Config/Admin/GetAll.status":                 kivik.StatusForbidden,
		"Config/Admin/GetSection.sections":           []string{"chicken"},
		"Config/Admin/GetSection/chicken.status":     kivik.StatusForbidden,
		"Config/NoAuth/GetAll.status":                kivik.StatusUnauthorized,
		"Config/NoAuth/GetSection.sections":          []string{"chicken"},
		"Config/NoAuth/GetItem.items":                []string{"foo.bar"},
		"Config/NoAuth/GetSection.status":            kivik.StatusUnauthorized,
		"Config/NoAuth/GetItem.status":               kivik.StatusUnauthorized,
		"Config/RW/group/NoAuth/Set.status":          kivik.StatusUnauthorized,
		"Config/RW/group/Admin/Set.status":           kivik.StatusForbidden,
		"Config/RW/group/NoAuth/Delete.status":       kivik.StatusForbidden,
		"Config/RW/group/NoAuth/Delete/group.status": kivik.StatusUnauthorized,
		"Config/RW/group/Admin/Delete.status":        kivik.StatusForbidden,

		"CreateDB/RW/NoAuth.status":         kivik.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": kivik.StatusPreconditionFailed,

		"DestroyDB/RW/Admin/NonExistantDB.status":  kivik.StatusNotFound,
		"DestroyDB/RW/NoAuth/NonExistantDB.status": kivik.StatusNotFound,
		"DestroyDB/RW/NoAuth/ExistingDB.status":    kivik.StatusUnauthorized,

		"AllDocs/Admin.databases":            []string{"_replicator", "chicken"},
		"AllDocs/Admin/_replicator.expected": []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":   0,
		"AllDocs/Admin/chicken.status":       kivik.StatusNotFound,
		"AllDocs/NoAuth.databases":           []string{"_replicator", "chicken"},
		"AllDocs/NoAuth/_replicator.status":  kivik.StatusUnauthorized,
		"AllDocs/NoAuth/chicken.status":      kivik.StatusNotFound,
		"AllDocs/RW/group/NoAuth.status":     kivik.StatusUnauthorized,

		"DBExists.databases":              []string{"_users", "chicken"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/NoAuth/_users.status":   kivik.StatusUnauthorized,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.status": kivik.StatusUnauthorized,

		"Membership.all_min_count":     2,
		"Membership.cluster_min_count": 2,

		"UUIDs.counts":                []int{-1, 0, 1, 10},
		"UUIDs/Admin/-1Count.status":  kivik.StatusBadRequest,
		"UUIDs/NoAuth/-1Count.status": kivik.StatusBadRequest,

		"Log/Admin.status":  kivik.StatusForbidden,
		"Log/NoAuth.status": kivik.StatusUnauthorized,

		"ServerInfo.version":        `^2\.0\.0$`,
		"ServerInfo.vendor":         `^IBM Cloudant$`,
		"ServerInfo.vendor_version": `^\d\d\d\d$`,

		"Get/RW/group/NoAuth/bob.status":   kivik.StatusUnauthorized,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusUnauthorized,
		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,
	})
}
