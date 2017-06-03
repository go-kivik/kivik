package test

import (
	"net/http"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/test/kt"
)

func init() {
	RegisterSuite(SuiteKivikServer, kt.SuiteConfig{
		"AllDBs.expected": []string{},
		"AllDBs/RW.skip":  true, // FIXME: Enable this when it's possible to delete DB from the server

		"CreateDB/RW.skip": true, // FIXME: Update when the server can destroy databases
		// "CreateDB/NoAuth.status":         http.StatusUnauthorized,
		// "CreateDB/Admin/Recreate.status": http.StatusPreconditionFailed,

		"DestroyDB.skip": true, // FIXME: Update when the server can destroy databases

		"AllDocs/Admin.databases":   []string{"foo"},
		"AllDocs/Admin/foo.status":  http.StatusNotFound,
		"AllDocs/RW.skip":           true, // FIXME: Update when the server can destroy databases
		"AllDocs/NoAuth.databases":  []string{"foo"},
		"AllDocs/NoAuth/foo.status": http.StatusNotFound,

		"DBExists.databases":            []string{"chicken"},
		"DBExists/Admin/chicken.exists": false,
		"DBExists/RW.skip":              true, // FIXME: Update when the server can destroy databases
		// "DBExists/RW/Admin.exists":      true,
		"DBExists/NoAuth.skip": true, // TODO

		"Log/Admin/Offset-1000.status":        http.StatusBadRequest,
		"Log/Admin/HTTP/TextBytes.status":     http.StatusBadRequest,
		"Log/Admin/HTTP/NegativeBytes.status": http.StatusBadRequest,
		"Log/NoAuth/Offset-1000.status":       http.StatusBadRequest,

		"Version.version":        `^0\.0\.0$`,
		"Version.vendor":         "Kivik",
		"Version.vendor_version": `^0\.0\.1$`,

		"Get.skip": true, // FIXME: Fix this when we can delete database

		"Put.skip": true, // FIXME: Fix this when we can write docs

		"Flush.databases":                     []string{"chicken"},
		"Flush/Admin/chicken/DoFlush.status":  kivik.StatusNotFound, // FIXME: Update when implemented
		"Flush/NoAuth/chicken/DoFlush.status": kivik.StatusNotFound, // FIXME: Update when implemented

		"Delete.skip": true, // FIXME: Fix this when we can delete docs.

		"Session/Get/Admin.info.authentication_handlers":  "default,cookie",
		"Session/Get/Admin.info.authentication_db":        "",
		"Session/Get/Admin.info.authenticated":            "cookie",
		"Session/Get/Admin.userCtx.roles":                 "_admin",
		"Session/Get/Admin.ok":                            "true",
		"Session/Get/NoAuth.info.authentication_handlers": "default,cookie",
		"Session/Get/NoAuth.info.authentication_db":       "",
		"Session/Get/NoAuth.info.authenticated":           "",
		"Session/Get/NoAuth.userCtx.roles":                "",
		"Session/Get/NoAuth.ok":                           "true",

		"Session/Post/EmptyJSON.status":                               kivik.StatusBadRequest,
		"Session/Post/BogusTypeJSON.status":                           kivik.StatusBadRequest,
		"Session/Post/BogusTypeForm.status":                           kivik.StatusBadRequest,
		"Session/Post/EmptyForm.status":                               kivik.StatusBadRequest,
		"Session/Post/BadJSON.status":                                 kivik.StatusBadRequest,
		"Session/Post/BadForm.status":                                 kivik.StatusBadRequest,
		"Session/Post/MeaninglessJSON.status":                         kivik.StatusBadRequest,
		"Session/Post/MeaninglessForm.status":                         kivik.StatusBadRequest,
		"Session/Post/GoodJSON.status":                                kivik.StatusUnauthorized,
		"Session/Post/BadQueryParam.status":                           kivik.StatusUnauthorized,
		"Session/Post/BadCredsJSON.status":                            kivik.StatusUnauthorized,
		"Session/Post/BadCredsForm.status":                            kivik.StatusUnauthorized,
		"Session/Post/GoodCredsJSONRemoteRedirAbsolute.status":        kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRedirRelativeNoSlash.status":       kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRemoteRedirHeaderInjection.status": kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRemoteRedirInvalidURL.status":      kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRedirEmpty.status":                 kivik.StatusBadRequest,
		"Session/Post/GoodCredsJSONRedirSchemaless.status":            kivik.StatusBadRequest,

		"Stats.skip":             true, // FIXME: Unimplemented
		"CreateDoc.skip":         true, // FIXME: Unimplemented
		"Compact.skip":           true, // FIXME: Unimplemented
		"ViewCleanup.skip":       true, // FIXME: Unimplemented
		"Security.skip":          true, // FIXME: Unimplemented
		"SetSecurity.skip":       true, // FIXME: Unimplemented
		"Rev.skip":               true, // FIXME: When Get works
		"DBUpdates.skip":         true, // FIXME: Unimplemented
		"Changes.skip":           true, // FIXME: Unimplemented
		"Copy.skip":              true, // FIXME: Unimplemented, depends on Get/Put or Copy
		"BulkDocs.skip":          true, // FIXME: Unimplemented
		"GetAttachment.skip":     true, // FIXME: Unimplemented
		"GetAttachmentMeta.skip": true, // FIXME: Unimplemented
		"PutAttachment.skip":     true, // FIXME: Unimplemented
		"DeleteAttachment.skip":  true, // FIXME: Unimplemented
		"Query.skip":             true, // FIXME: Unimplemented
		"Find.skip":              true, // FIXME: Unimplemented
		"CreateIndex.skip":       true, // FIXME: Unimplemented
		"GetIndexes.skip":        true, // FIXME: Unimplemented
		"DeleteIndex.skip":       true, // FIXME: Unimplemented
		"GetReplications.skip":   true, // FIXME: Unimplemented
		"Replicate.skip":         true, // FIXME: Unimplemented
	})
}
