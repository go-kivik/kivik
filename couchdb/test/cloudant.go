// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package test

import (
	"net/http"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

// nolint:gomnd
func registerSuiteCloudant() {
	kiviktest.RegisterSuite(kiviktest.SuiteCloudant, kt.SuiteConfig{
		"Options":                       httpClient(),
		"AllDBs.expected":               []string{"_replicator", "_users"},
		"AllDBs/NoAuth.status":          http.StatusUnauthorized,
		"AllDBs/RW/group/NoAuth.status": http.StatusUnauthorized,

		"CreateDB/RW/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": http.StatusPreconditionFailed,

		"DestroyDB/RW/Admin/NonExistantDB.status":  http.StatusNotFound,
		"DestroyDB/RW/NoAuth/NonExistantDB.status": http.StatusNotFound,
		"DestroyDB/RW/NoAuth/ExistingDB.status":    http.StatusUnauthorized,

		"AllDocs.databases":                  []string{"_replicator", "chicken", "_duck"},
		"AllDocs/Admin/_replicator.expected": []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":   0,
		"AllDocs/Admin/chicken.status":       http.StatusNotFound,
		"AllDocs/Admin/_duck.status":         http.StatusForbidden,
		"AllDocs/NoAuth/_replicator.status":  http.StatusUnauthorized,
		"AllDocs/NoAuth/chicken.status":      http.StatusNotFound,
		"AllDocs/NoAuth/_duck.status":        http.StatusUnauthorized,
		"AllDocs/RW/group/NoAuth.status":     http.StatusUnauthorized,

		"Find.databases":                         []string{"_replicator", "chicken", "_duck"},
		"Find/Admin/_replicator.expected":        []string{"_design/_replicator"},
		"Find/Admin/_replicator.offset":          0,
		"Find/Admin/chicken.status":              http.StatusNotFound,
		"Find/Admin/_duck.status":                http.StatusForbidden,
		"Find/NoAuth/_replicator.status":         http.StatusUnauthorized,
		"Find/NoAuth/chicken.status":             http.StatusNotFound,
		"Find/NoAuth/_duck.status":               http.StatusUnauthorized,
		"Find/RW/group/NoAuth.status":            http.StatusUnauthorized,
		"Find/Admin/_replicator/Warning.warning": "no matching index found, create an index to optimize query time",
		"Find/RW/group/Admin/Warning.warning":    "no matching index found, create an index to optimize query time",
		"Find/RW/group/NoAuth/Warning.warning":   "no matching index found, create an index to optimize query time",

		"Explain.databases":             []string{"chicken", "_duck"},
		"Explain/Admin/chicken.status":  http.StatusNotFound,
		"Explain/Admin/_duck.status":    http.StatusForbidden,
		"Explain/NoAuth/chicken.status": http.StatusNotFound,
		"Explain/NoAuth/_duck.status":   http.StatusUnauthorized,

		"Query/RW/group/NoAuth.status": http.StatusUnauthorized,

		"DBExists.databases":              []string{"_users", "chicken", "_duck"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/Admin/_duck.status":     http.StatusForbidden,
		"DBExists/NoAuth/_users.status":   http.StatusUnauthorized,
		"DBExists/NoAuth/_duck.status":    http.StatusUnauthorized,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.status": http.StatusUnauthorized,

		"Log/Admin.status":              http.StatusForbidden,
		"Log/NoAuth.status":             http.StatusUnauthorized,
		"Log/Admin/Offset-1000.status":  http.StatusBadRequest,
		"Log/NoAuth/Offset-1000.status": http.StatusBadRequest,

		"Version.version":        `^2\.0\.0$`,
		"Version.vendor":         `^IBM Cloudant$`,
		"Version.vendor_version": `^\d\d\d\d$`,

		"Get/RW/group/Admin/bogus.status":        http.StatusNotFound,
		"Get/RW/group/NoAuth/bob.status":         http.StatusUnauthorized,
		"Get/RW/group/NoAuth/bogus.status":       http.StatusUnauthorized,
		"Get/RW/group/NoAuth/_design/foo.status": http.StatusUnauthorized,
		"Get/RW/group/NoAuth/_local/foo.status":  http.StatusUnauthorized,

		"GetRev/RW/group/Admin/bogus.status":        http.StatusNotFound,
		"GetRev/RW/group/NoAuth/bob.status":         http.StatusUnauthorized,
		"GetRev/RW/group/NoAuth/bogus.status":       http.StatusUnauthorized,
		"GetRev/RW/group/NoAuth/_design/foo.status": http.StatusUnauthorized,
		"GetRev/RW/group/NoAuth/_local/foo.status":  http.StatusUnauthorized,

		"Put/RW/NoAuth/Create.status": http.StatusUnauthorized,

		"Flush.databases":                     []string{"_users", "chicken", "_duck"},
		"Flush/Admin/chicken/DoFlush.status":  http.StatusNotFound,
		"Flush/Admin/_duck/DoFlush.status":    http.StatusForbidden,
		"Flush/NoAuth/chicken/DoFlush.status": http.StatusNotFound,
		"Flush/NoAuth/_users/DoFlush.status":  http.StatusUnauthorized,
		"Flush/NoAuth/_duck/DoFlush.status":   http.StatusUnauthorized,

		"Delete/RW/Admin/group/MissingDoc.status":       http.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status": http.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":         http.StatusConflict,
		"Delete/RW/NoAuth.status":                       http.StatusUnauthorized,

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

		"Session/Post/EmptyJSON.status":                               http.StatusBadRequest,
		"Session/Post/BogusTypeJSON.status":                           http.StatusBadRequest,
		"Session/Post/BogusTypeForm.status":                           http.StatusBadRequest,
		"Session/Post/EmptyForm.status":                               http.StatusBadRequest,
		"Session/Post/BadJSON.status":                                 http.StatusBadRequest,
		"Session/Post/BadForm.status":                                 http.StatusBadRequest,
		"Session/Post/MeaninglessJSON.status":                         http.StatusInternalServerError,
		"Session/Post/MeaninglessForm.status":                         http.StatusBadRequest,
		"Session/Post/GoodJSON.status":                                http.StatusUnauthorized,
		"Session/Post/BadQueryParam.status":                           http.StatusUnauthorized,
		"Session/Post/BadCredsJSON.status":                            http.StatusUnauthorized,
		"Session/Post/BadCredsForm.status":                            http.StatusUnauthorized,
		"Session/Post/GoodCredsJSONRemoteRedirHeaderInjection.status": http.StatusBadRequest,
		"Session/Post/GoodCredsJSONRemoteRedirInvalidURL.skip":        true, // Cloudant doesn't sanitize the Location value, so sends unparseable headers.

		"Stats.databases":             []string{"_users", "chicken", "_duck"},
		"Stats/Admin/chicken.status":  http.StatusNotFound,
		"Stats/Admin/_duck.status":    http.StatusForbidden,
		"Stats/NoAuth/_users.status":  http.StatusUnauthorized,
		"Stats/NoAuth/chicken.status": http.StatusNotFound,
		"Stats/NoAuth/_duck.status":   http.StatusUnauthorized,
		"Stats/RW/NoAuth.status":      http.StatusUnauthorized,

		"CreateDoc/RW/group/NoAuth.status": http.StatusUnauthorized,

		"Compact/RW/Admin.status":  http.StatusForbidden,
		"Compact/RW/NoAuth.status": http.StatusUnauthorized,

		"ViewCleanup/RW/Admin.status":  http.StatusForbidden,
		"ViewCleanup/RW/NoAuth.status": http.StatusUnauthorized,

		"Security.databases":                    []string{"_replicator", "_users", "_global_changes", "chicken", "_duck"},
		"Security/Admin/_global_changes.status": http.StatusForbidden,
		"Security/Admin/chicken.status":         http.StatusNotFound,
		"Security/Admin/_duck.status":           http.StatusForbidden,
		"Security/NoAuth.status":                http.StatusUnauthorized,
		"Security/NoAuth/chicken.status":        http.StatusNotFound,
		"Security/NoAuth/_duck.status":          http.StatusUnauthorized,
		"Security/RW/group/NoAuth.status":       http.StatusUnauthorized,

		"SetSecurity/RW/Admin/NotExists.status":  http.StatusNotFound,
		"SetSecurity/RW/NoAuth/NotExists.status": http.StatusNotFound,
		"SetSecurity/RW/NoAuth/Exists.status":    http.StatusUnauthorized,

		"DBUpdates/RW/Admin.status":  http.StatusNotFound, // Cloudant apparently disables this
		"DBUpdates/RW/NoAuth.status": http.StatusUnauthorized,

		"Changes/RW/group/NoAuth.status": http.StatusUnauthorized,

		"Copy/RW/group/NoAuth.status": http.StatusUnauthorized,

		"BulkDocs/RW/NoAuth.status":                    http.StatusUnauthorized,
		"BulkDocs/RW/NoAuth/group/Mix/Conflict.status": http.StatusConflict,
		"BulkDocs/RW/Admin/group/Mix/Conflict.status":  http.StatusConflict,

		"GetAttachment/RW/group/Admin/foo/NotFound.status": http.StatusNotFound,
		"GetAttachment/RW/group/NoAuth.status":             http.StatusUnauthorized,

		"GetAttachmentMeta/RW/group/Admin/foo/NotFound.status": http.StatusNotFound,
		"GetAttachmentMeta/RW/group/NoAuth.status":             http.StatusUnauthorized,

		"PutAttachment/RW/group/Admin/Conflict.status": http.StatusInternalServerError, // COUCHDB-3361
		"PutAttachment/RW/group/NoAuth.status":         http.StatusUnauthorized,

		// "DeleteAttachment/RW/group/Admin/NotFound.status":  http.StatusNotFound, // COUCHDB-3362
		"DeleteAttachment/RW/group/NoAuth.status":       http.StatusUnauthorized,
		"DeleteAttachment/RW/group/Admin/NoDoc.status":  http.StatusInternalServerError,
		"DeleteAttachment/RW/group/NoAuth/NoDoc.status": http.StatusUnauthorized,

		"Put/RW/Admin/group/LeadingUnderscoreInID.status": http.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":              http.StatusConflict,
		"Put/RW/NoAuth/group.status":                      http.StatusUnauthorized,
		"Put/RW/NoAuth/group/Conflict.skip":               true,

		"CreateIndex/RW/Admin/group/EmptyIndex.status":    http.StatusBadRequest,
		"CreateIndex/RW/Admin/group/BlankIndex.status":    http.StatusBadRequest,
		"CreateIndex/RW/Admin/group/InvalidIndex.status":  http.StatusBadRequest,
		"CreateIndex/RW/Admin/group/NilIndex.status":      http.StatusBadRequest,
		"CreateIndex/RW/Admin/group/InvalidJSON.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/EmptyIndex.status":   http.StatusUnauthorized,
		"CreateIndex/RW/NoAuth/group/BlankIndex.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/InvalidIndex.status": http.StatusUnauthorized,
		"CreateIndex/RW/NoAuth/group/NilIndex.status":     http.StatusUnauthorized,
		"CreateIndex/RW/NoAuth/group/InvalidJSON.status":  http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/Valid.status":        http.StatusUnauthorized,

		"GetIndexes.databases":                     []string{"_replicator", "_users", "_global_changes", "chicken", "_duck"},
		"GetIndexes/Admin/_replicator.indexes":     []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/_users.indexes":          []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/_global_changes.status":  http.StatusForbidden,
		"GetIndexes/Admin/chicken.status":          http.StatusNotFound,
		"GetIndexes/Admin/_duck.status":            http.StatusForbidden,
		"GetIndexes/NoAuth/_replicator.indexes":    []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_users.indexes":         []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_global_changes.status": http.StatusForbidden,
		"GetIndexes/NoAuth/chicken.status":         http.StatusNotFound,
		"GetIndexes/NoAuth/_duck.status":           http.StatusForbidden,
		"GetIndexes/RW/NoAuth.status":              http.StatusUnauthorized,

		"DeleteIndex/RW/Admin/group/NotFoundDdoc.status": http.StatusNotFound,
		"DeleteIndex/RW/Admin/group/NotFoundName.status": http.StatusNotFound,
		"DeleteIndex/RW/NoAuth.status":                   http.StatusUnauthorized,

		"GetReplications/NoAuth.status": http.StatusUnauthorized,

		"Replicate.NotFoundDB":                                  "http://localhost:5984/foo",
		"Replicate.timeoutSeconds":                              300,
		"Replicate/RW/NoAuth.status":                            http.StatusUnauthorized,
		"Replicate/RW/Admin/group/MissingSource/Results.status": http.StatusInternalServerError,
		"Replicate/RW/Admin/group/MissingTarget/Results.status": http.StatusInternalServerError,

		"Query/RW/group/Admin/WithoutDocs/ScanDoc.status":  http.StatusBadRequest,
		"Query/RW/group/NoAuth/WithoutDocs/ScanDoc.status": http.StatusBadRequest,
	})
}
