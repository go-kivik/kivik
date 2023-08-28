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
func registerSuiteCouch31() {
	kiviktest.RegisterSuite(kiviktest.SuiteCouch31, kt.SuiteConfig{
		"Options":                       httpClient(),
		"AllDBs.expected":               []string{"_global_changes", "_replicator", "_users"},
		"AllDBs/RW/group/NoAuth.status": http.StatusUnauthorized,
		"AllDBs/NoAuth.status":          http.StatusUnauthorized,

		"CreateDB/RW/NoAuth.status":                  http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status":          http.StatusPreconditionFailed,
		"CreateDoc/RW/group/NoAuth/WithID.status":    http.StatusUnauthorized,
		"CreateDoc/RW/group/NoAuth/WithoutID.status": http.StatusUnauthorized,

		"DestroyDB/RW/NoAuth.status":              http.StatusUnauthorized,
		"DestroyDB/RW/Admin/NonExistantDB.status": http.StatusNotFound,

		"AllDocs.databases":                          []string{"chicken", "_duck"},
		"AllDocs/Admin/_replicator.offset":           0,
		"AllDocs/Admin/chicken.status":               http.StatusNotFound,
		"AllDocs/Admin/_duck.status":                 http.StatusNotFound,
		"AllDocs/NoAuth/_replicator.status":          http.StatusUnauthorized,
		"AllDocs/NoAuth/chicken.status":              http.StatusNotFound,
		"AllDocs/NoAuth/_duck.status":                http.StatusUnauthorized,
		"AllDocs/RW/group/NoAuth/WithDocs.status":    http.StatusUnauthorized,
		"AllDocs/RW/group/NoAuth/WithoutDocs.status": http.StatusUnauthorized,

		"Find.databases":                      []string{"chicken", "_duck"},
		"Find/Admin/chicken.status":           http.StatusNotFound,
		"Find/Admin/_duck.status":             http.StatusNotFound,
		"Find/NoAuth/chicken.status":          http.StatusNotFound,
		"Find/NoAuth/_duck.status":            http.StatusUnauthorized,
		"Find/RW/group/Admin/Warning.warning": "No matching index found, create an index to optimize query time.",
		"Find/RW/group/NoAuth.status":         http.StatusUnauthorized,

		"Explain.databases":              []string{"chicken", "_duck"},
		"Explain/Admin/chicken.status":   http.StatusNotFound,
		"Explain/Admin/_duck.status":     http.StatusNotFound,
		"Explain/NoAuth/chicken.status":  http.StatusNotFound,
		"Explain/NoAuth/_duck.status":    http.StatusUnauthorized,
		"Explain/RW/group/NoAuth.status": http.StatusUnauthorized,
		"Explain.plan": &kivik.QueryPlan{
			Index: map[string]interface{}{
				"ddoc": nil,
				"name": "_all_docs",
				"type": "special",
				"def":  map[string]interface{}{"fields": []interface{}{map[string]string{"_id": "asc"}}},
			},
			Selector: map[string]interface{}{"_id": map[string]interface{}{"$gt": nil}},
			Options: map[string]interface{}{
				"bookmark":        "nil",
				"conflicts":       false,
				"execution_stats": false,
				"r":               []int{49},
				"sort":            map[string]interface{}{},
				"use_index":       []interface{}{},
				"stable":          false,
				"stale":           false,
				"update":          true,
				"skip":            0,
				"limit":           25,
				"partition":       "",
				"fields":          "all_fields",
			},
			Range: nil,
			Limit: 25,
		},

		"DBExists.databases":              []string{"_users", "chicken", "_duck"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/Admin/_duck.exists":     false,
		"DBExists/NoAuth/_users.status":   http.StatusUnauthorized,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/NoAuth/_duck.status":    http.StatusUnauthorized,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.status": http.StatusUnauthorized,

		"Log.skip": true, // This was removed in CouchDB 2.0

		"Version.version":        `^3\.1\.`,
		"Version.vendor":         `^The Apache Software Foundation$`,
		"Version.vendor_version": ``, // CouchDB 2.0+ no longer has a vendor version

		"Get/RW/group/Admin/bogus.status":        http.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status":       http.StatusUnauthorized,
		"Get/RW/group/NoAuth/_design/foo.status": http.StatusUnauthorized,
		"Get/RW/group/NoAuth/bob.status":         http.StatusUnauthorized,
		"Get/RW/group/NoAuth/_local/foo.status":  http.StatusUnauthorized,

		"GetRev/RW/group/Admin/bogus.status":        http.StatusNotFound,
		"GetRev/RW/group/NoAuth/bogus.status":       http.StatusUnauthorized,
		"GetRev/RW/group/NoAuth/_local/foo.status":  http.StatusUnauthorized,
		"GetRev/RW/group/NoAuth/_design/foo.status": http.StatusUnauthorized,
		"GetRev/RW/group/NoAuth/bob.status":         http.StatusUnauthorized,

		"Flush.databases":                     []string{"_users", "chicken", "_duck"},
		"Flush/NoAuth/chicken/DoFlush.status": http.StatusNotFound,
		"Flush/Admin/chicken/DoFlush.status":  http.StatusNotFound,
		"Flush/Admin/_duck/DoFlush.status":    http.StatusNotFound,
		"Flush/NoAuth/_duck/DoFlush.status":   http.StatusUnauthorized,
		"Flush/NoAuth/_users/DoFlush.status":  http.StatusUnauthorized,

		"Delete/RW/Admin/group/MissingDoc.status":        http.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status":  http.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":          http.StatusConflict,
		"Delete/RW/NoAuth/group/MissingDoc.status":       http.StatusUnauthorized,
		"Delete/RW/NoAuth/group/InvalidRevFormat.status": http.StatusUnauthorized,
		"Delete/RW/NoAuth/group/WrongRev.status":         http.StatusUnauthorized,
		"Delete/RW/NoAuth/group/DesignDoc.status":        http.StatusUnauthorized,
		"Delete/RW/NoAuth/group/ValidRev.status":         http.StatusUnauthorized,
		"Delete/RW/NoAuth/group/Local.status":            http.StatusUnauthorized,

		"Session/Get/Admin.info.authentication_handlers":  "cookie,default",
		"Session/Get/Admin.info.authentication_db":        "",
		"Session/Get/Admin.info.authenticated":            "cookie",
		"Session/Get/Admin.userCtx.roles":                 "_admin",
		"Session/Get/Admin.ok":                            "true",
		"Session/Get/NoAuth.info.authentication_handlers": "cookie,default",
		"Session/Get/NoAuth.info.authentication_db":       "",
		"Session/Get/NoAuth.info.authenticated":           "",
		"Session/Get/NoAuth.userCtx.roles":                "",
		"Session/Get/NoAuth.ok":                           "true",

		"Session/Post/EmptyJSON.status":                             http.StatusBadRequest,
		"Session/Post/BogusTypeJSON.status":                         http.StatusBadRequest,
		"Session/Post/BogusTypeForm.status":                         http.StatusBadRequest,
		"Session/Post/EmptyForm.status":                             http.StatusBadRequest,
		"Session/Post/BadJSON.status":                               http.StatusBadRequest,
		"Session/Post/BadForm.status":                               http.StatusBadRequest,
		"Session/Post/MeaninglessJSON.status":                       http.StatusInternalServerError,
		"Session/Post/MeaninglessForm.status":                       http.StatusBadRequest,
		"Session/Post/GoodJSON.status":                              http.StatusUnauthorized,
		"Session/Post/BadQueryParam.status":                         http.StatusUnauthorized,
		"Session/Post/BadCredsJSON.status":                          http.StatusUnauthorized,
		"Session/Post/BadCredsForm.status":                          http.StatusUnauthorized,
		"Session/Post/GoodCredsJSONRemoteRedirAbsolute.status":      http.StatusBadRequest, // As of 2.1.1 all redirect paths must begin with '/'
		"Session/Post/GoodCredsJSONRedirEmpty.status":               http.StatusBadRequest, // As of 2.1.1 all redirect paths must begin with '/'
		"Session/Post/GoodCredsJSONRedirRelativeNoSlash.status":     http.StatusBadRequest, // As of 2.1.1 all redirect paths must begin with '/'
		"Session/Post/GoodCredsJSONRemoteRedirHeaderInjection.skip": true,                  // CouchDB allows header injection
		"Session/Post/GoodCredsJSONRemoteRedirInvalidURL.skip":      true,                  // CouchDB doesn't sanitize the Location value, so sends unparseable headers.

		"Stats.databases":             []string{"_users", "chicken", "_duck"},
		"Stats/Admin/chicken.status":  http.StatusNotFound,
		"Stats/Admin/_duck.status":    http.StatusNotFound,
		"Stats/NoAuth/chicken.status": http.StatusNotFound,
		"Stats/NoAuth/_duck.status":   http.StatusUnauthorized,
		"Stats/NoAuth/_users.status":  http.StatusUnauthorized,
		"Stats/RW/NoAuth.status":      http.StatusUnauthorized,

		"Compact.skip":             false,
		"Compact/RW/NoAuth.status": http.StatusUnauthorized,

		"Security.databases":                     []string{"_replicator", "_users", "_global_changes", "chicken", "_duck"},
		"Security/Admin/chicken.status":          http.StatusNotFound,
		"Security/Admin/_duck.status":            http.StatusNotFound,
		"Security/NoAuth/_global_changes.status": http.StatusUnauthorized,
		"Security/NoAuth/chicken.status":         http.StatusNotFound,
		"Security/NoAuth/_duck.status":           http.StatusUnauthorized,
		"Security/RW/group/NoAuth.status":        http.StatusUnauthorized,
		"SetSecurity/RW/Admin/NotExists.status":  http.StatusNotFound,
		"SetSecurity/RW/NoAuth/NotExists.status": http.StatusNotFound,
		"SetSecurity/RW/NoAuth/Exists.status":    http.StatusUnauthorized,
		"Security/NoAuth/_replicator.status":     http.StatusUnauthorized,
		"Security/NoAuth/_users.status":          http.StatusUnauthorized,

		"DBUpdates/RW/NoAuth.status": http.StatusUnauthorized,

		"BulkDocs/RW/NoAuth/group/Mix/Conflict.status": http.StatusConflict,
		"BulkDocs/RW/Admin/group/Mix/Conflict.status":  http.StatusConflict,
		"BulkDocs/RW/NoAuth/group/Mix.status":          http.StatusUnauthorized,
		"BulkDocs/RW/NoAuth/group/Delete.status":       http.StatusUnauthorized,
		"BulkDocs/RW/NoAuth/group/Update.status":       http.StatusUnauthorized,
		"BulkDocs/RW/NoAuth/group/Create.status":       http.StatusUnauthorized,
		"BulkDocs/RW/NoAuth/group/NonJSON.status":      http.StatusUnauthorized,

		"GetAttachment/RW/group/Admin/foo/NotFound.status":         http.StatusNotFound,
		"GetAttachment/RW/group/NoAuth/foo/NotFound.status":        http.StatusUnauthorized,
		"GetAttachment/RW/group/NoAuth/_design/foo/foo.txt.status": http.StatusUnauthorized,
		"GetAttachment/RW/group/NoAuth/foo/foo.txt.status":         http.StatusUnauthorized,

		"GetAttachmentMeta/RW/group/Admin/foo/NotFound.status":         http.StatusNotFound,
		"GetAttachmentMeta/RW/group/NoAuth/foo/NotFound.status":        http.StatusUnauthorized,
		"GetAttachmentMeta/RW/group/NoAuth/_design/foo/foo.txt.status": http.StatusUnauthorized,
		"GetAttachmentMeta/RW/group/NoAuth/foo/foo.txt.status":         http.StatusUnauthorized,

		"PutAttachment/RW/group/Admin/Conflict.status":         http.StatusConflict,
		"PutAttachment/RW/group/NoAuth/Conflict.status":        http.StatusUnauthorized,
		"PutAttachment/RW/group/NoAuth/UpdateDesignDoc.status": http.StatusUnauthorized,
		"PutAttachment/RW/group/NoAuth/CreateDesignDoc.status": http.StatusUnauthorized,
		"PutAttachment/RW/group/NoAuth/Update.status":          http.StatusUnauthorized,
		"PutAttachment/RW/group/NoAuth/Create.status":          http.StatusUnauthorized,

		// "DeleteAttachment/RW/group/Admin/NotFound.status":  http.StatusNotFound, // COUCHDB-3362
		"DeleteAttachment/RW/group/NoAuth/NotFound.status":  http.StatusUnauthorized,
		"DeleteAttachment/RW/group/Admin/NoDoc.status":      http.StatusConflict,
		"DeleteAttachment/RW/group/NoAuth/NoDoc.status":     http.StatusUnauthorized,
		"DeleteAttachment/RW/group/NoAuth/DesignDoc.status": http.StatusUnauthorized,
		"DeleteAttachment/RW/group/NoAuth/foo.txt.status":   http.StatusUnauthorized,

		"Put/RW/Admin/group/LeadingUnderscoreInID.status":  http.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":               http.StatusConflict,
		"Put/RW/NoAuth/group/LeadingUnderscoreInID.status": http.StatusUnauthorized,
		"Put/RW/NoAuth/group/DesignDoc.status":             http.StatusUnauthorized,
		"Put/RW/NoAuth/group/Conflict.status":              http.StatusUnauthorized,
		"Put/RW/NoAuth/group/HeavilyEscapedID":             http.StatusUnauthorized,
		"Put/RW/NoAuth/group/Local.status":                 http.StatusUnauthorized,
		"Put/RW/NoAuth/group/HeavilyEscapedID.status":      http.StatusUnauthorized,
		"Put/RW/NoAuth/group/SlashInID.status":             http.StatusUnauthorized,
		"Put/RW/NoAuth/group/Create.status":                http.StatusUnauthorized,

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
		"GetIndexes/Admin/_global_changes.indexes": []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/chicken.status":          http.StatusNotFound,
		"GetIndexes/Admin/_duck.status":            http.StatusNotFound,
		"GetIndexes/NoAuth/_replicator.indexes":    []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_users.indexes":         []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_global_changes.status": http.StatusUnauthorized,
		"GetIndexes/NoAuth/chicken.status":         http.StatusNotFound,
		"GetIndexes/NoAuth/_duck.status":           http.StatusUnauthorized,
		"GetIndexes/NoAuth/_replicator.status":     http.StatusUnauthorized,
		"GetIndexes/NoAuth/_users.status":          http.StatusUnauthorized,
		"GetIndexes/RW.indexes": []kivik.Index{
			kt.AllDocsIndex,
			{
				DesignDoc: "_design/foo",
				Name:      "bar",
				Type:      "json",
				Definition: map[string]interface{}{
					"fields": []map[string]string{
						{"foo": "asc"},
					},
					"partial_filter_selector": map[string]string{},
				},
			},
		},

		"DeleteIndex/RW/Admin/group/NotFoundDdoc.status":  http.StatusNotFound,
		"DeleteIndex/RW/Admin/group/NotFoundName.status":  http.StatusNotFound,
		"DeleteIndex/RW/NoAuth/group/NotFoundDdoc.status": http.StatusUnauthorized,
		"DeleteIndex/RW/NoAuth/group/NotFoundName.status": http.StatusUnauthorized,
		"DeleteIndex/RW/NoAuth/group/ValidIndex.status":   http.StatusUnauthorized,

		"GetReplications/NoAuth.status": http.StatusUnauthorized,

		"Replicate.NotFoundDB":                                  "http://localhost:5984/foo",
		"Replicate.timeoutSeconds":                              60,
		"Replicate.prefix":                                      "http://localhost:5984/",
		"Replicate/RW/Admin.prefix":                             "http://admin:abc123@localhost:5984/",
		"Replicate/RW/NoAuth.status":                            http.StatusUnauthorized,
		"Replicate/RW/Admin/group/MissingSource/Results.status": http.StatusNotFound,
		"Replicate/RW/Admin/group/MissingTarget/Results.status": http.StatusNotFound,

		"Query/RW/group/Admin/WithoutDocs/ScanDoc.status": http.StatusBadRequest,
		"Query/RW/group/NoAuth/WithDocs.status":           http.StatusUnauthorized,
		"Query/RW/group/NoAuth/WithoutDocs.status":        http.StatusUnauthorized,

		"ViewCleanup/RW/NoAuth.status": http.StatusUnauthorized,

		"Changes/Continuous/RW/group/NoAuth.status": http.StatusUnauthorized,
		"Changes/Normal/RW/group/NoAuth.status":     http.StatusUnauthorized,
		"Changes/Continuous.options": map[string]interface{}{
			"feed":      "continuous",
			"since":     "now",
			"heartbeat": 6000,
		},

		"Copy/RW/group/NoAuth/_design/foo.status": http.StatusUnauthorized,
		"Copy/RW/group/NoAuth/foo.status":         http.StatusUnauthorized,
		"Copy/RW/group/NoAuth/_local/foo.status":  http.StatusUnauthorized,
	})
}
