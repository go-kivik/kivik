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

func registerSuiteCouch22() {
	kiviktest.RegisterSuite(kiviktest.SuiteCouch22, kt.SuiteConfig{
		"Options":         httpClient(),
		"AllDBs.expected": []string{"_global_changes", "_replicator", "_users"},

		"CreateDB/RW/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": http.StatusPreconditionFailed,

		"DestroyDB/RW/NoAuth.status":              http.StatusUnauthorized,
		"DestroyDB/RW/Admin/NonExistantDB.status": http.StatusNotFound,

		"AllDocs.databases":                 []string{"chicken", "_duck"},
		"AllDocs/Admin/_replicator.offset":  0,
		"AllDocs/Admin/chicken.status":      http.StatusNotFound,
		"AllDocs/Admin/_duck.status":        http.StatusNotFound,
		"AllDocs/NoAuth/_replicator.status": http.StatusUnauthorized,
		"AllDocs/NoAuth/chicken.status":     http.StatusNotFound,
		"AllDocs/NoAuth/_duck.status":       http.StatusUnauthorized,

		"Find.databases":                 []string{"chicken", "_duck"},
		"Find/Admin/chicken.status":      http.StatusNotFound,
		"Find/Admin/_duck.status":        http.StatusNotFound,
		"Find/NoAuth/chicken.status":     http.StatusNotFound,
		"Find/NoAuth/_duck.status":       http.StatusUnauthorized,
		"Find/RW/Admin/Warning.warning":  "no matching index found, create an index to optimize query time",
		"Find/RW/NoAuth/Warning.warning": "no matching index found, create an index to optimize query time",

		"Explain.databases":             []string{"chicken", "_duck"},
		"Explain/Admin/chicken.status":  http.StatusNotFound,
		"Explain/Admin/_duck.status":    http.StatusNotFound,
		"Explain/NoAuth/chicken.status": http.StatusNotFound,
		"Explain/NoAuth/_duck.status":   http.StatusUnauthorized,
		"Explain.plan": &kivik.QueryPlan{
			Index: map[string]any{
				"ddoc": nil,
				"name": "_all_docs",
				"type": "special",
				"def":  map[string]any{"fields": []any{map[string]string{"_id": "asc"}}},
			},
			Selector: map[string]any{"_id": map[string]any{"$gt": nil}},
			Options: map[string]any{
				"bookmark":        "nil",
				"conflicts":       false,
				"execution_stats": false,
				"r":               []int{49},
				"sort":            map[string]any{},
				"use_index":       []any{},
				"stable":          false,
				"stale":           false,
				"update":          true,
				"skip":            0,
				"limit":           25,
				"fields":          "all_fields",
			},
			Range: nil,
			Limit: 25,
		},

		"DBExists.databases":             []string{"_users", "chicken", "_duck"},
		"DBExists/Admin/_users.exists":   true,
		"DBExists/Admin/chicken.exists":  false,
		"DBExists/Admin/_duck.exists":    false,
		"DBExists/NoAuth/_users.exists":  true,
		"DBExists/NoAuth/chicken.exists": false,
		"DBExists/NoAuth/_duck.status":   http.StatusUnauthorized,
		"DBExists/RW/Admin.exists":       true,
		"DBExists/RW/NoAuth.exists":      true,

		// "DBsStats/NoAuth.status": http.StatusUnauthorized,

		// "AllDBsStats/NoAuth.status": http.StatusUnauthorized,

		"Log.skip": true, // This was removed in CouchDB 2.0

		"Version.version":        `^2\.2\.`,
		"Version.vendor":         `^The Apache Software Foundation$`,
		"Version.vendor_version": ``, // CouchDB 2.0+ no longer has a vendor version

		"Get/RW/Admin/bogus.status":  http.StatusNotFound,
		"Get/RW/NoAuth/bogus.status": http.StatusNotFound,

		"GetRev/RW/Admin/bogus.status":  http.StatusNotFound,
		"GetRev/RW/NoAuth/bogus.status": http.StatusNotFound,

		"Flush.databases":                     []string{"_users", "chicken", "_duck"},
		"Flush/NoAuth/chicken/DoFlush.status": http.StatusNotFound,
		"Flush/Admin/chicken/DoFlush.status":  http.StatusNotFound,
		// "Flush/Admin/_duck/DoFlush.status":    http.StatusNotFound, // Possible bug: https://github.com/apache/couchdb/issues/1585
		"Flush/NoAuth/_duck/DoFlush.status": http.StatusUnauthorized,

		"Delete/RW/Admin/MissingDoc.status":        http.StatusNotFound,
		"Delete/RW/Admin/InvalidRevFormat.status":  http.StatusBadRequest,
		"Delete/RW/Admin/WrongRev.status":          http.StatusConflict,
		"Delete/RW/NoAuth/MissingDoc.status":       http.StatusNotFound,
		"Delete/RW/NoAuth/InvalidRevFormat.status": http.StatusBadRequest,
		"Delete/RW/NoAuth/WrongRev.status":         http.StatusConflict,
		"Delete/RW/NoAuth/DesignDoc.status":        http.StatusUnauthorized,

		"Session/Get/Admin.info.authentication_handlers":  "cookie,default",
		"Session/Get/Admin.info.authentication_db":        "_users",
		"Session/Get/Admin.info.authenticated":            "cookie",
		"Session/Get/Admin.userCtx.roles":                 "_admin",
		"Session/Get/Admin.ok":                            "true",
		"Session/Get/NoAuth.info.authentication_handlers": "cookie,default",
		"Session/Get/NoAuth.info.authentication_db":       "_users",
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

		"Compact.skip":             false,
		"Compact/RW/NoAuth.status": http.StatusUnauthorized,

		"Security.databases":                     []string{"_replicator", "_users", "_global_changes", "chicken", "_duck"},
		"Security/Admin/chicken.status":          http.StatusNotFound,
		"Security/Admin/_duck.status":            http.StatusNotFound,
		"Security/NoAuth/_global_changes.status": http.StatusUnauthorized,
		"Security/NoAuth/chicken.status":         http.StatusNotFound,
		"Security/NoAuth/_duck.status":           http.StatusUnauthorized,
		"Security/RW/NoAuth.status":              http.StatusUnauthorized,

		"SetSecurity/RW/Admin/NotExists.status":  http.StatusNotFound,
		"SetSecurity/RW/NoAuth/NotExists.status": http.StatusNotFound,
		"SetSecurity/RW/NoAuth/Exists.status":    http.StatusInternalServerError, // Can you say WTF?

		"DBUpdates/RW/NoAuth.status": http.StatusUnauthorized,

		"BulkDocs/RW/NoAuth/Mix/Conflict.status": http.StatusConflict,
		"BulkDocs/RW/Admin/Mix/Conflict.status":  http.StatusConflict,

		"GetAttachment/RW/Admin/foo/NotFound.status":  http.StatusNotFound,
		"GetAttachment/RW/NoAuth/foo/NotFound.status": http.StatusNotFound,

		"GetAttachmentMeta/RW/Admin/foo/NotFound.status":  http.StatusNotFound,
		"GetAttachmentMeta/RW/NoAuth/foo/NotFound.status": http.StatusNotFound,

		"PutAttachment/RW/Admin/Conflict.status":         http.StatusConflict,
		"PutAttachment/RW/NoAuth/Conflict.status":        http.StatusConflict,
		"PutAttachment/RW/NoAuth/UpdateDesignDoc.status": http.StatusUnauthorized,
		"PutAttachment/RW/NoAuth/CreateDesignDoc.status": http.StatusUnauthorized,

		// "DeleteAttachment/RW/Admin/NotFound.status":  http.StatusNotFound, // COUCHDB-3362
		// "DeleteAttachment/RW/NoAuth/NotFound.status": http.StatusNotFound, // COUCHDB-3362
		"DeleteAttachment/RW/Admin/NoDoc.status":      http.StatusConflict,
		"DeleteAttachment/RW/NoAuth/NoDoc.status":     http.StatusConflict,
		"DeleteAttachment/RW/NoAuth/DesignDoc.status": http.StatusUnauthorized,

		"Put/RW/Admin/LeadingUnderscoreInID.status":  http.StatusBadRequest,
		"Put/RW/Admin/Conflict.status":               http.StatusConflict,
		"Put/RW/NoAuth/LeadingUnderscoreInID.status": http.StatusBadRequest,
		"Put/RW/NoAuth/DesignDoc.status":             http.StatusUnauthorized,
		"Put/RW/NoAuth/Conflict.status":              http.StatusConflict,

		"CreateIndex/RW/Admin/EmptyIndex.status":    http.StatusBadRequest,
		"CreateIndex/RW/Admin/BlankIndex.status":    http.StatusBadRequest,
		"CreateIndex/RW/Admin/InvalidIndex.status":  http.StatusBadRequest,
		"CreateIndex/RW/Admin/NilIndex.status":      http.StatusBadRequest,
		"CreateIndex/RW/Admin/InvalidJSON.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/EmptyIndex.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/BlankIndex.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/InvalidIndex.status": http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/NilIndex.status":     http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/InvalidJSON.status":  http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/Valid.status":        http.StatusInternalServerError, // COUCHDB-3374

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
		"GetIndexes/RW.indexes": []kivik.Index{
			kt.AllDocsIndex,
			{
				DesignDoc: "_design/foo",
				Name:      "bar",
				Type:      "json",
				Definition: map[string]any{
					"fields": []map[string]string{
						{"foo": "asc"},
					},
					"partial_filter_selector": map[string]string{},
				},
			},
		},

		"DeleteIndex/RW/Admin/NotFoundDdoc.status":  http.StatusNotFound,
		"DeleteIndex/RW/Admin/NotFoundName.status":  http.StatusNotFound,
		"DeleteIndex/RW/NoAuth/NotFoundDdoc.status": http.StatusNotFound,
		"DeleteIndex/RW/NoAuth/NotFoundName.status": http.StatusNotFound,

		"GetReplications/NoAuth.status": http.StatusUnauthorized,

		"Replicate.NotFoundDB":                            "http://localhost:5984/foo",
		"Replicate.timeoutSeconds":                        60,
		"Replicate.prefix":                                "http://localhost:5984/",
		"Replicate/RW/NoAuth.status":                      http.StatusForbidden,
		"Replicate/RW/Admin/MissingSource/Results.status": http.StatusNotFound,
		"Replicate/RW/Admin/MissingTarget/Results.status": http.StatusNotFound,

		"Query/RW/Admin/WithoutDocs/ScanDoc.status":  http.StatusBadRequest,
		"Query/RW/NoAuth/WithoutDocs/ScanDoc.status": http.StatusBadRequest,

		"ViewCleanup/RW/NoAuth.status": http.StatusUnauthorized,

		"Changes/Continuous.options": kivik.Params(map[string]any{
			"feed":      "continuous",
			"since":     "now",
			"heartbeat": 6000,
		}),
	})
}
