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

		"Log/Admin.status":              kivik.StatusForbidden,
		"Log/NoAuth.status":             kivik.StatusUnauthorized,
		"Log/Admin/Offset-1000.status":  kivik.StatusBadRequest,
		"Log/NoAuth/Offset-1000.status": kivik.StatusBadRequest,

		"ServerInfo.version":        `^2\.0\.0$`,
		"ServerInfo.vendor":         `^IBM Cloudant$`,
		"ServerInfo.vendor_version": `^\d\d\d\d$`,

		"Get/RW/group/NoAuth/bob.status":   kivik.StatusUnauthorized,
		"Get/RW/group/NoAuth/bogus.status": kivik.StatusUnauthorized,
		"Get/RW/group/Admin/bogus.status":  kivik.StatusNotFound,

		"Put/RW/NoAuth/Create.status": kivik.StatusUnauthorized,

		"Flush.databases":                           []string{"_users", "chicken"},
		"Flush/Admin/chicken/DoFlush.status":        kivik.StatusNotFound,
		"Flush/NoAuth/chicken/DoFlush.status":       kivik.StatusNotFound,
		"Flush/NoAuth/_users/DoFlush.status":        kivik.StatusUnauthorized,
		"Flush/Admin/_users/DoFlush/Timestamp.skip": true, // Cloudant always sets the timestamp to 0

		"Delete/RW/Admin/group/MissingDoc.status":       kivik.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status": kivik.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":         kivik.StatusConflict,
		"Delete/RW/NoAuth.status":                       kivik.StatusUnauthorized,

		"Session/Get/Admin.info.authentication_handlers":  "delegated,cookie,default,local",
		"Session/Get/Admin.info.authentication_db":        "_users",
		"Session/Get/Admin.info.authenticated":            "cookie",
		"Session/Get/Admin.userCtx.roles":                 "_admin,_reader,_writer",
		"Session/Get/Admin.ok":                            "true",
		"Session/Get/NoAuth.info.authentication_handlers": "delegated,cookie,default,local",
		"Session/Get/NoAuth.info.authentication_db":       "_users",
		"Session/Get/NoAuth.info.authenticated":           "local",
		"Session/Get/NoAuth.userCtx.roles":                "",
		"Session/Get/NoAuth.ok":                           "true",

		"Session/Post/EmptyJSON.status":                               kivik.StatusBadRequest,
		"Session/Post/BogusTypeJSON.status":                           kivik.StatusBadRequest,
		"Session/Post/BogusTypeForm.status":                           kivik.StatusBadRequest,
		"Session/Post/EmptyForm.status":                               kivik.StatusBadRequest,
		"Session/Post/BadJSON.status":                                 kivik.StatusBadRequest,
		"Session/Post/BadForm.status":                                 kivik.StatusBadRequest,
		"Session/Post/MeaninglessJSON.status":                         kivik.StatusInternalServerError,
		"Session/Post/MeaninglessForm.status":                         kivik.StatusBadRequest,
		"Session/Post/GoodJSON.status":                                kivik.StatusUnauthorized,
		"Session/Post/BadQueryParam.status":                           kivik.StatusUnauthorized,
		"Session/Post/BadCredsJSON.status":                            kivik.StatusUnauthorized,
		"Session/Post/BadCredsForm.status":                            kivik.StatusUnauthorized,
		"Session/Post/GoodCredsJSONRemoteRedirHeaderInjection.status": kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRemoteRedirInvalidURL.skip":        true, // Cloudant doesn't sanitize the Location value, so sends unparseable headers.

		"DBInfo.databases":        []string{"_users"},
		"DBInfo/NoAuth.status":    kivik.StatusUnauthorized,
		"DBInfo/RW/NoAuth.status": kivik.StatusUnauthorized,

		"CreateDoc/RW/group/NoAuth.status": kivik.StatusUnauthorized,

		"Compact/RW/Admin.status":  kivik.StatusForbidden,
		"Compact/RW/NoAuth.status": kivik.StatusUnauthorized,

		"ViewCleanup/RW/Admin.status":  kivik.StatusForbidden,
		"ViewCleanup/RW/NoAuth.status": kivik.StatusUnauthorized,
	})
}
