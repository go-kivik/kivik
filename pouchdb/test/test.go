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

//go:build js

// Package test provides PouchDB integration tests.
package test

import (
	"net/http"
	"net/url"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/gopherjs/gopherjs/js"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	if pouchDB := js.Global.Get("PouchDB"); pouchDB != js.Undefined {
		memPouch := js.Global.Get("PouchDB").Call("defaults", map[string]interface{}{
			"db": js.Global.Call("require", "memdown"),
		})
		js.Global.Set("PouchDB", memPouch)
	}
}

// RegisterPouchDBSuites registers the PouchDB test suites.
func RegisterPouchDBSuites() {
	kiviktest.RegisterSuite(kiviktest.SuitePouchLocal, kt.SuiteConfig{
		"PreCleanup.skip": true,

		// Features which are not supported by PouchDB
		"Log.skip":         true,
		"Flush.skip":       true,
		"Security.skip":    true, // FIXME: Perhaps implement later with a plugin?
		"SetSecurity.skip": true, // FIXME: Perhaps implement later with a plugin?
		"DBUpdates.skip":   true,

		"AllDBs.skip":   true, // FIXME: Find a way to test with the plugin
		"CreateDB.skip": true, // FIXME: No way to validate if this works unless/until allDbs works
		"DBExists.skip": true, // FIXME: Maybe fix this if/when allDBs works?

		"AllDocs/Admin.databases":                        []string{},
		"AllDocs/RW/group/Admin/WithDocs/UpdateSeq.skip": true,

		"Find/Admin.databases": []string{},
		// TODO: Fix this and uncomment https://github.com/go-kivik/kivik/issues/588
		// "Find/RW/group/Admin/Warning.warning": indexWarning,

		"Explain.databases": []string{},
		"Explain.plan": &kivik.QueryPlan{
			Index: map[string]interface{}{
				"ddoc": nil,
				"name": "_all_docs",
				"type": "special",
				"def":  map[string]interface{}{"fields": []interface{}{map[string]string{"_id": "asc"}}},
			},
			Selector: map[string]interface{}{"_id": map[string]interface{}{"$gt": nil}},
			Options: map[string]interface{}{
				"bookmark":  "nil",
				"conflicts": false,
				"r":         []int{49},
				"sort":      map[string]interface{}{},
				"use_index": []interface{}{},
			},
			Fields: func() []interface{} {
				if ver := runtime.Version(); strings.HasPrefix(ver, "go1.16") {
					return []interface{}{}
				}
				// From GopherJS 17 on, null arrays are properly converted to nil
				return nil
			}(),
			Range: map[string]interface{}{
				"start_key": nil,
			},
		},

		"Query/RW/group/Admin/WithDocs/UpdateSeq.skip": true,

		"Version.version":        `^[789]\.\d\.\d$`,
		"Version.vendor":         `^PouchDB$`,
		"Version.vendor_version": `^[789]\.\d\.\d$`,

		"Get/RW/group/Admin/bogus.status":  http.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": http.StatusNotFound,

		"GetRev/RW/group/Admin/bogus.status":  http.StatusNotFound,
		"GetRev/RW/group/NoAuth/bogus.status": http.StatusNotFound,

		"Delete/RW/Admin/group/MissingDoc.status":       http.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status": http.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":         http.StatusConflict,

		"Stats/Admin.skip": true, // No predefined DBs for Local PouchDB

		"BulkDocs/RW/Admin/group/Mix/Conflict.status": http.StatusConflict,

		"GetAttachment/RW/group/Admin/foo/NotFound.status": http.StatusNotFound,

		"GetAttachmentMeta/RW/group/Admin/foo/NotFound.status": http.StatusNotFound,

		"PutAttachment/RW/group/Admin/Conflict.status": http.StatusConflict,

		// "DeleteAttachment/RW/group/Admin/NotFound.status": http.StatusNotFound, // https://github.com/pouchdb/pouchdb/issues/6409
		"DeleteAttachment/RW/group/Admin/NoDoc.status": http.StatusNotFound,

		"Put/RW/Admin/group/LeadingUnderscoreInID.status": http.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":              http.StatusConflict,

		"CreateIndex/RW/Admin/group/EmptyIndex.status":   http.StatusInternalServerError,
		"CreateIndex/RW/Admin/group/BlankIndex.status":   http.StatusBadRequest,
		"CreateIndex/RW/Admin/group/InvalidIndex.status": http.StatusInternalServerError,
		"CreateIndex/RW/Admin/group/NilIndex.status":     http.StatusInternalServerError,
		"CreateIndex/RW/Admin/group/InvalidJSON.status":  http.StatusBadRequest,

		"GetIndexes.databases": []string{},

		"DeleteIndex/RW/Admin/group/NotFoundDdoc.status": http.StatusNotFound,
		"DeleteIndex/RW/Admin/group/NotFoundName.status": http.StatusNotFound,

		"Replicate.skip": true, // No need to do this for both Local and Remote

		"Query/RW/group/Admin/WithoutDocs/ScanDoc.status": http.StatusBadRequest,

		"Changes/Continuous.options": kivik.Params(map[string]interface{}{
			"live":    true,
			"timeout": false,
		}),
	})
	kiviktest.RegisterSuite(kiviktest.SuitePouchRemote, kt.SuiteConfig{
		// Features which are not supported by PouchDB
		"Log.skip":         true,
		"Flush.skip":       true,
		"Session.skip":     true,
		"Security.skip":    true, // FIXME: Perhaps implement later with a plugin?
		"SetSecurity.skip": true, // FIXME: Perhaps implement later with a plugin?
		"DBUpdates.skip":   true,

		"PreCleanup.skip": true,

		"AllDBs.skip": true, // FIXME: Perhaps a workaround can be found?

		"CreateDB/RW/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": http.StatusPreconditionFailed,

		"DBExists.databases":              []string{"_users", "chicken"},
		"DBExists/Admin/_users.exists":    true,
		"DBExists/Admin/chicken.exists":   false,
		"DBExists/NoAuth/_users.exists":   true,
		"DBExists/NoAuth/chicken.exists":  false,
		"DBExists/RW/group/Admin.exists":  true,
		"DBExists/RW/group/NoAuth.exists": true,

		"DestroyDB/RW/NoAuth/NonExistantDB.status": http.StatusNotFound,
		"DestroyDB/RW/Admin/NonExistantDB.status":  http.StatusNotFound,
		"DestroyDB/RW/NoAuth/ExistingDB.status":    http.StatusUnauthorized,

		"AllDocs.databases":                                  []string{"_replicator", "_users", "chicken"},
		"AllDocs/Admin/_replicator.expected":                 []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":                   0,
		"AllDocs/Admin/_users.expected":                      []string{"_design/_auth"},
		"AllDocs/Admin/chicken.status":                       http.StatusNotFound,
		"AllDocs/NoAuth/_replicator.status":                  http.StatusUnauthorized,
		"AllDocs/NoAuth/_users.status":                       http.StatusUnauthorized,
		"AllDocs/NoAuth/chicken.status":                      http.StatusNotFound,
		"AllDocs/Admin/_replicator/WithDocs/UpdateSeq.skip":  true,
		"AllDocs/Admin/_users/WithDocs/UpdateSeq.skip":       true,
		"AllDocs/RW/group/Admin/WithDocs/UpdateSeq.skip":     true,
		"AllDocs/RW/group/Admin/WithoutDocs/UpdateSeq.skip":  true,
		"AllDocs/RW/group/NoAuth/WithDocs/UpdateSeq.skip":    true,
		"AllDocs/RW/group/NoAuth/WithoutDocs/UpdateSeq.skip": true,

		"Find.databases":             []string{"chicken", "_duck"},
		"Find/Admin/chicken.status":  http.StatusNotFound,
		"Find/Admin/_duck.status":    http.StatusNotFound,
		"Find/NoAuth/chicken.status": http.StatusNotFound,
		"Find/NoAuth/_duck.status":   http.StatusUnauthorized,
		// TODO: Fix this and uncomment https://github.com/go-kivik/kivik/issues/588
		// "Find/RW/group/Admin/Warning.warning":  "No matching index found, create an index to optimize query time",
		"Find/RW/group/NoAuth/Warning.warning": "No matching index found, create an index to optimize query time",

		"Explain.databases":             []string{"chicken", "_duck"},
		"Explain/Admin/chicken.status":  http.StatusNotFound,
		"Explain/Admin/_duck.status":    http.StatusNotFound,
		"Explain/NoAuth/chicken.status": http.StatusNotFound,
		"Explain/NoAuth/_duck.status":   http.StatusUnauthorized,
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
				"limit":           25, //nolint:gomnd
				"fields":          "all_fields",
			},
			Fields: func() []interface{} {
				if ver := runtime.Version(); strings.HasPrefix(ver, "go1.16") {
					return []interface{}{}
				}
				// From GopherJS 17 on, null arrays are properly converted to nil
				return nil
			}(),
			Range: nil,
			Limit: 25, //nolint:gomnd
		},

		"CreateIndex/RW/Admin/group/EmptyIndex.status":    http.StatusBadRequest,
		"CreateIndex/RW/Admin/group/BlankIndex.status":    http.StatusBadRequest,
		"CreateIndex/RW/Admin/group/InvalidIndex.status":  http.StatusBadRequest,
		"CreateIndex/RW/Admin/group/NilIndex.status":      http.StatusBadRequest,
		"CreateIndex/RW/Admin/group/InvalidJSON.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/EmptyIndex.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/BlankIndex.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/InvalidIndex.status": http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/NilIndex.status":     http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/InvalidJSON.status":  http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/group/Valid.status":        http.StatusInternalServerError, // COUCHDB-3374

		"GetIndexes.databases":                     []string{"_replicator", "_users", "_global_changes"},
		"GetIndexes/Admin/_replicator.indexes":     []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/_users.indexes":          []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/_global_changes.indexes": []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_replicator.indexes":    []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_users.indexes":         []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_global_changes.skip":   true, // Pouch connects to the DB before searching the Index, so this test fails
		"GetIndexes/NoAuth/_global_changes.status": http.StatusUnauthorized,
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
		"DeleteIndex/RW/NoAuth/group/NotFoundDdoc.status": http.StatusNotFound,
		"DeleteIndex/RW/NoAuth/group/NotFoundName.status": http.StatusNotFound,

		"Query/RW/group/Admin/WithDocs/UpdateSeq.skip":  true,
		"Query/RW/group/NoAuth/WithDocs/UpdateSeq.skip": true,

		"Version.version":        `^[789]\.\d\.\d$`,
		"Version.vendor":         `^PouchDB$`,
		"Version.vendor_version": `^[789]\.\d\.\d$`,

		"Get/RW/group/Admin/bogus.status":  http.StatusNotFound,
		"Get/RW/group/NoAuth/bogus.status": http.StatusNotFound,

		"GetRev/RW/group/Admin/bogus.status":  http.StatusNotFound,
		"GetRev/RW/group/NoAuth/bogus.status": http.StatusNotFound,

		"Delete/RW/Admin/group/MissingDoc.status":        http.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status":  http.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":          http.StatusConflict,
		"Delete/RW/NoAuth/group/MissingDoc.status":       http.StatusNotFound,
		"Delete/RW/NoAuth/group/InvalidRevFormat.status": http.StatusBadRequest,
		"Delete/RW/NoAuth/group/WrongRev.status":         http.StatusConflict,
		"Delete/RW/NoAuth/group/DesignDoc.status":        http.StatusUnauthorized,

		"Stats.databases":             []string{"_users", "chicken"},
		"Stats/Admin/chicken.status":  http.StatusNotFound,
		"Stats/NoAuth/chicken.status": http.StatusNotFound,

		"BulkDocs/RW/NoAuth/group/Mix/Conflict.status": http.StatusConflict,
		"BulkDocs/RW/Admin/group/Mix/Conflict.status":  http.StatusConflict,

		"GetAttachment/RW/group/Admin/foo/NotFound.status":  http.StatusNotFound,
		"GetAttachment/RW/group/NoAuth/foo/NotFound.status": http.StatusNotFound,

		"GetAttachmentMeta/RW/group/Admin/foo/NotFound.status":  http.StatusNotFound,
		"GetAttachmentMeta/RW/group/NoAuth/foo/NotFound.status": http.StatusNotFound,

		"PutAttachment/RW/group/Admin/Conflict.status":         http.StatusConflict,
		"PutAttachment/RW/group/NoAuth/Conflict.status":        http.StatusConflict,
		"PutAttachment/RW/group/NoAuth/UpdateDesignDoc.status": http.StatusUnauthorized,
		"PutAttachment/RW/group/NoAuth/CreateDesignDoc.status": http.StatusUnauthorized,

		// "DeleteAttachment/RW/group/Admin/NotFound.status":  http.StatusNotFound, // COUCHDB-3362
		// "DeleteAttachment/RW/group/NoAuth/NotFound.status": http.StatusNotFound, // COUCHDB-3362
		"DeleteAttachment/RW/group/Admin/NoDoc.status":      http.StatusConflict,
		"DeleteAttachment/RW/group/NoAuth/NoDoc.status":     http.StatusConflict,
		"DeleteAttachment/RW/group/NoAuth/DesignDoc.status": http.StatusUnauthorized,

		"Put/RW/Admin/group/LeadingUnderscoreInID.status":  http.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":               http.StatusConflict,
		"Put/RW/NoAuth/group/DesignDoc.status":             http.StatusUnauthorized,
		"Put/RW/NoAuth/group/LeadingUnderscoreInID.status": http.StatusBadRequest,
		"Put/RW/NoAuth/group/Conflict.status":              http.StatusConflict,

		"Replicate.NotFoundDB": func() string {
			var dsn string
			for _, env := range []string{
				"KIVIK_TESt_DSN_COUCH22", "KIVIK_TEST_DSN_COUCH21",
				"KIVIK_TEST_DSN_COUCH20", "KIVIK_TEST_DSN_COUCH16",
				"KIVIK_TEST_DSN_CLOUDANT",
			} {
				dsn = os.Getenv(env)
				if dsn != "" {
					break
				}
			}
			parsed, _ := url.Parse(dsn)
			parsed.User = nil
			return strings.TrimSuffix(parsed.String(), "/") + "/doesntexist"
		}(),
		"Replicate.prefix":         "none",
		"Replicate.timeoutSeconds": 5, //nolint:gomnd
		"Replicate.mode":           "pouchdb",
		"Replicate/RW/Admin/group/MissingSource/Results.status":  http.StatusUnauthorized,
		"Replicate/RW/Admin/group/MissingTarget/Results.status":  http.StatusUnauthorized,
		"Replicate/RW/NoAuth/group/MissingSource/Results.status": http.StatusUnauthorized,
		"Replicate/RW/NoAuth/group/MissingTarget/Results.status": http.StatusUnauthorized,

		"Query/RW/group/Admin/WithoutDocs/ScanDoc.status":  http.StatusBadRequest,
		"Query/RW/group/NoAuth/WithoutDocs/ScanDoc.status": http.StatusBadRequest,

		// "ViewCleanup/RW/NoAuth.status": http.StatusUnauthorized, # FIXME: #14

		"Changes/Continuous.options": kivik.Params(map[string]interface{}{
			"live":    true,
			"timeout": false,
		}),
	})
}

// PouchLocalTest runs the PouchDB tests against a local database.
func PouchLocalTest(t *testing.T) {
	client, err := kivik.New("pouch", "")
	if err != nil {
		t.Errorf("Failed to connect to PouchDB driver: %s", err)
		return
	}
	clients := &kt.Context{
		RW:    true,
		Admin: client,
		T:     t,
	}
	kiviktest.RunTestsInternal(clients, kiviktest.SuitePouchLocal)
}

// PouchRemoteTest runs the PouchDB tests against a remote CouchDB database.
func PouchRemoteTest(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuitePouchRemote, "KIVIK_TEST_DSN_COUCH22")
}
