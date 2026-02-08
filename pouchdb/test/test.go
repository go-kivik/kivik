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
	"github.com/go-kivik/kivik/v4/pouchdb/internal"
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

		"AllDocs/Admin.databases":                  []string{},
		"AllDocs/RW/Admin/WithDocs/UpdateSeq.skip": true,

		"Find/Admin.databases": []string{},
		// TODO: Fix this and uncomment https://github.com/go-kivik/kivik/issues/588
		// "Find/RW/Admin/Warning.warning": indexWarning,

		"Explain.databases": []string{},
		"Explain.plan": &kivik.QueryPlan{
			Index: map[string]interface{}{
				"ddoc": nil,
				"name": "_all_docs",
				"type": "special",
				"def":  map[string]interface{}{"fields": []interface{}{map[string]interface{}{"_id": "asc"}}},
			},
			Selector: map[string]interface{}{"_id": map[string]interface{}{"$gt": nil}},
			Limit: func() int64 {
				if strings.HasPrefix(internal.MustPouchDBVersion(), "9.") {
					return 25
				}
				return 0
			}(),
			Options: func() map[string]interface{} {
				options := map[string]interface{}{
					"bookmark":  "nil",
					"conflicts": false,
					"r":         []int{49},
					"sort":      map[string]interface{}{},
					"use_index": []interface{}{},
				}
				if strings.HasPrefix(internal.MustPouchDBVersion(), "9.") {
					options["limit"] = float64(25)
				}
				return options
			}(),
			Fields: nil,
			Range: map[string]interface{}{
				"start_key": nil,
			},
		},

		"Query/RW/Admin/WithDocs/UpdateSeq.skip": true,

		"Version.version":        `^[789]\.\d\.\d$`,
		"Version.vendor":         `^PouchDB$`,
		"Version.vendor_version": `^[789]\.\d\.\d$`,

		"Get/RW/Admin/bogus.status":  http.StatusNotFound,
		"Get/RW/NoAuth/bogus.status": http.StatusNotFound,

		"GetRev/RW/Admin/bogus.status":  http.StatusNotFound,
		"GetRev/RW/NoAuth/bogus.status": http.StatusNotFound,

		"Delete/RW/Admin/MissingDoc.status":       http.StatusNotFound,
		"Delete/RW/Admin/InvalidRevFormat.status": http.StatusBadRequest,
		"Delete/RW/Admin/WrongRev.status":         http.StatusConflict,

		"Stats/Admin.skip": true, // No predefined DBs for Local PouchDB

		"BulkDocs/RW/Admin/Mix/Conflict.status": http.StatusConflict,

		"GetAttachment/RW/Admin/foo/NotFound.status": http.StatusNotFound,

		"GetAttachmentMeta/RW/Admin/foo/NotFound.status": http.StatusNotFound,

		"PutAttachment/RW/Admin/Conflict.status": http.StatusConflict,

		// "DeleteAttachment/RW/Admin/NotFound.status": http.StatusNotFound, // https://github.com/pouchdb/pouchdb/issues/6409
		"DeleteAttachment/RW/Admin/NoDoc.status": http.StatusNotFound,

		"Put/RW/Admin/LeadingUnderscoreInID.status": http.StatusBadRequest,
		"Put/RW/Admin/Conflict.status":              http.StatusConflict,

		"CreateIndex/RW/Admin/EmptyIndex.status":   http.StatusInternalServerError,
		"CreateIndex/RW/Admin/BlankIndex.status":   http.StatusBadRequest,
		"CreateIndex/RW/Admin/InvalidIndex.status": http.StatusInternalServerError,
		"CreateIndex/RW/Admin/NilIndex.status":     http.StatusInternalServerError,
		"CreateIndex/RW/Admin/InvalidJSON.status":  http.StatusBadRequest,

		"GetIndexes.databases": []string{},

		"DeleteIndex/RW/Admin/NotFoundDdoc.status": http.StatusNotFound,
		"DeleteIndex/RW/Admin/NotFoundName.status": http.StatusNotFound,

		"Replicate.skip": true, // No need to do this for both Local and Remote

		"Query/RW/Admin/WithoutDocs/ScanDoc.status": http.StatusBadRequest,

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

		"AllDBs.skip":      true, // FIXME: Perhaps a workaround can be found?
		"AllDBsStats.skip": true, // FIXME: Depends on AllDBs

		"CreateDB/RW/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": http.StatusPreconditionFailed,

		"DBExists.databases":             []string{"_users", "chicken"},
		"DBExists/Admin/_users.exists":   true,
		"DBExists/Admin/chicken.exists":  false,
		"DBExists/NoAuth/_users.exists":  true,
		"DBExists/NoAuth/chicken.exists": false,
		"DBExists/RW/Admin.exists":       true,
		"DBExists/RW/NoAuth.exists":      true,
		"DBExists/NoAuth/_users.status":  http.StatusUnauthorized,
		"DBExists/RW/NoAuth.status":      http.StatusUnauthorized,

		"DestroyDB/RW/NoAuth/NonExistantDB.status": http.StatusNotFound,
		"DestroyDB/RW/Admin/NonExistantDB.status":  http.StatusNotFound,
		"DestroyDB/RW/NoAuth/ExistingDB.status":    http.StatusUnauthorized,

		"AllDocs.databases": []string{"_replicator", "_users", "chicken"},
		// "AllDocs/Admin/_replicator.expected":                 []string{"_design/_replicator"},
		"AllDocs/Admin/_replicator.offset":                  0,
		"AllDocs/Admin/_users.expected":                     []string{"_design/_auth"},
		"AllDocs/NoAuth/_replicator.status":                 http.StatusUnauthorized,
		"AllDocs/NoAuth/_users.status":                      http.StatusUnauthorized,
		"AllDocs/NoAuth/chicken.status":                     http.StatusUnauthorized,
		"AllDocs/Admin/_replicator/WithDocs/UpdateSeq.skip": true,
		"AllDocs/Admin/_users/WithDocs/UpdateSeq.skip":      true,
		"AllDocs/RW/Admin/WithDocs/UpdateSeq.skip":          true,
		"AllDocs/RW/Admin/WithoutDocs/UpdateSeq.skip":       true,
		"AllDocs/RW/NoAuth/WithDocs.status":                 http.StatusUnauthorized,
		"AllDocs/RW/NoAuth/WithoutDocs.status":              http.StatusUnauthorized,
		"AllDocs/RW/NoAuth/WithDocs/UpdateSeq.skip":         true,
		"AllDocs/RW/NoAuth/WithoutDocs/UpdateSeq.skip":      true,
		"AllDocs/Admin/chicken/WithDocs/UpdateSeq.skip":     true,

		"Find.databases":                     []string{"chicken", "_duck"},
		"Find/Admin/chicken/Warning.warning": "No matching index found, create an index to optimize query time.",
		"Find/Admin/_duck.status":            http.StatusBadRequest,
		"Find/NoAuth/chicken.status":         http.StatusUnauthorized,
		"Find/NoAuth/_duck.status":           http.StatusUnauthorized,
		"Find/RW/Admin/Warning.warning":      "No matching index found, create an index to optimize query time.",
		"Find/RW/NoAuth/Warning.warning":     "No matching index found, create an index to optimize query time.",

		"Explain.databases":             []string{"chicken", "_duck"},
		"Explain/Admin/_duck.status":    http.StatusBadRequest,
		"Explain/NoAuth/chicken.status": http.StatusUnauthorized,
		"Explain/NoAuth/_duck.status":   http.StatusUnauthorized,
		"Explain.plan": &kivik.QueryPlan{
			Index: map[string]interface{}{
				"ddoc": nil,
				"name": "_all_docs",
				"type": "special",
				"def":  map[string]interface{}{"fields": []interface{}{map[string]interface{}{"_id": "asc"}}},
			},
			Selector: map[string]interface{}{"_id": map[string]interface{}{"$gt": nil}},
			Options: map[string]interface{}{
				"bookmark":        "nil",
				"conflicts":       false,
				"execution_stats": false,
				"partition":       "",
				"r":               []interface{}{float64(49)},
				"sort":            map[string]interface{}{},
				"use_index":       []interface{}{},
				"stable":          false,
				"stale":           false,
				"update":          true,
				"skip":            0,
				"limit":           float64(25),
				"fields":          "all_fields",
			},
			Fields: nil,
			Range:  nil,
			Limit:  25,
		},

		"CreateIndex/RW/Admin/EmptyIndex.status":    http.StatusBadRequest,
		"CreateIndex/RW/Admin/BlankIndex.status":    http.StatusBadRequest,
		"CreateIndex/RW/Admin/InvalidIndex.status":  http.StatusBadRequest,
		"CreateIndex/RW/Admin/NilIndex.status":      http.StatusBadRequest,
		"CreateIndex/RW/Admin/InvalidJSON.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/EmptyIndex.status":   http.StatusUnauthorized,
		"CreateIndex/RW/NoAuth/BlankIndex.status":   http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/InvalidIndex.status": http.StatusUnauthorized,
		"CreateIndex/RW/NoAuth/NilIndex.status":     http.StatusUnauthorized,
		"CreateIndex/RW/NoAuth/InvalidJSON.status":  http.StatusBadRequest,
		"CreateIndex/RW/NoAuth/Valid.status":        http.StatusUnauthorized,

		"GetIndexes.databases":                     []string{"_replicator", "_users", "_global_changes"},
		"GetIndexes/Admin/_replicator.indexes":     []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/_users.indexes":          []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/Admin/_global_changes.indexes": []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_replicator.indexes":    []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_replicator.status":     http.StatusUnauthorized,
		"GetIndexes/NoAuth/_users.indexes":         []kivik.Index{kt.AllDocsIndex},
		"GetIndexes/NoAuth/_users.status":          http.StatusUnauthorized,
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

		"DeleteIndex/RW/Admin/NotFoundDdoc.status":  http.StatusNotFound,
		"DeleteIndex/RW/Admin/NotFoundName.status":  http.StatusNotFound,
		"DeleteIndex/RW/NoAuth/NotFoundDdoc.status": http.StatusUnauthorized,
		"DeleteIndex/RW/NoAuth/NotFoundName.status": http.StatusUnauthorized,

		"Query/RW/Admin/WithDocs/UpdateSeq.skip":  true,
		"Query/RW/NoAuth/WithDocs/UpdateSeq.skip": true,

		"Version.version":        `^[789]\.\d\.\d$`,
		"Version.vendor":         `^PouchDB$`,
		"Version.vendor_version": `^[789]\.\d\.\d$`,

		"Get/RW/Admin/bogus.status":        http.StatusNotFound,
		"Get/RW/NoAuth/bob.status":         http.StatusUnauthorized,
		"Get/RW/NoAuth/_design/foo.status": http.StatusUnauthorized,
		"Get/RW/NoAuth/_local/foo.status":  http.StatusUnauthorized,
		"Get/RW/NoAuth/bogus.status":       http.StatusUnauthorized,

		"GetRev/RW/Admin/bogus.status":        http.StatusNotFound,
		"GetRev/RW/NoAuth/bob.status":         http.StatusUnauthorized,
		"GetRev/RW/NoAuth/foo.status":         http.StatusUnauthorized,
		"GetRev/RW/NoAuth/bogus.status":       http.StatusUnauthorized,
		"GetRev/RW/NoAuth/_design/foo.status": http.StatusUnauthorized,
		"GetRev/RW/NoAuth/_local/foo.status":  http.StatusUnauthorized,

		"Delete/RW/Admin/MissingDoc.status":        http.StatusNotFound,
		"Delete/RW/Admin/InvalidRevFormat.status":  http.StatusBadRequest,
		"Delete/RW/Admin/WrongRev.status":          http.StatusConflict,
		"Delete/RW/NoAuth/MissingDoc.status":       http.StatusUnauthorized,
		"Delete/RW/NoAuth/InvalidRevFormat.status": http.StatusUnauthorized,
		"Delete/RW/NoAuth/WrongRev.status":         http.StatusUnauthorized,
		"Delete/RW/NoAuth/DesignDoc.status":        http.StatusUnauthorized,
		"Delete/RW/NoAuth/Local.status":            http.StatusUnauthorized,
		"Delete/RW/NoAuth/ValidRev.status":         http.StatusUnauthorized,

		"Stats.databases":        []string{"_users", "chicken"},
		"Stats/NoAuth.status":    http.StatusUnauthorized,
		"Stats/RW/NoAuth.status": http.StatusUnauthorized,

		"BulkDocs/RW/NoAuth/Mix/Conflict.status": http.StatusConflict,
		"BulkDocs/RW/Admin/Mix/Conflict.status":  http.StatusConflict,
		"BulkDocs/RW/NoAuth/Create.status":       http.StatusUnauthorized,
		"BulkDocs/RW/NoAuth/Update.status":       http.StatusUnauthorized,
		"BulkDocs/RW/NoAuth/Delete.status":       http.StatusUnauthorized,
		"BulkDocs/RW/NoAuth/Mix.status":          http.StatusUnauthorized,
		"BulkDocs/RW/NoAuth/NonJSON.status":      http.StatusUnauthorized,

		"GetAttachment/RW/Admin/foo/NotFound.status":  http.StatusNotFound,
		"GetAttachment/RW/NoAuth/foo/NotFound.status": http.StatusUnauthorized,
		"GetAttachment/RW/NoAuth.status":              http.StatusUnauthorized,

		"GetAttachmentMeta/RW/Admin/foo/NotFound.status": http.StatusNotFound,
		"GetAttachmentMeta/RW/NoAuth.status":             http.StatusUnauthorized,

		"PutAttachment/RW/Admin/Conflict.status":         http.StatusConflict,
		"PutAttachment/RW/NoAuth/Create.status":          http.StatusUnauthorized,
		"PutAttachment/RW/NoAuth/Update.status":          http.StatusUnauthorized,
		"PutAttachment/RW/NoAuth/Conflict.status":        http.StatusUnauthorized,
		"PutAttachment/RW/NoAuth/UpdateDesignDoc.status": http.StatusUnauthorized,
		"PutAttachment/RW/NoAuth/CreateDesignDoc.status": http.StatusUnauthorized,

		"DeleteAttachment/RW/Admin/NotFound.status":  http.StatusNotFound,
		"DeleteAttachment/RW/NoAuth/NotFound.status": http.StatusUnauthorized,
		"DeleteAttachment/RW/Admin/NoDoc.status":     http.StatusConflict,
		"DeleteAttachment/RW/NoAuth.status":          http.StatusUnauthorized,

		"Put/RW/Admin/LeadingUnderscoreInID.status":  http.StatusBadRequest,
		"Put/RW/Admin/Conflict.status":               http.StatusConflict,
		"Put/RW/NoAuth/Create.status":                http.StatusUnauthorized,
		"Put/RW/NoAuth/DesignDoc.status":             http.StatusUnauthorized,
		"Put/RW/NoAuth/Local.status":                 http.StatusUnauthorized,
		"Put/RW/NoAuth/LeadingUnderscoreInID.status": http.StatusBadRequest,
		"Put/RW/NoAuth/HeavilyEscapedID.status":      http.StatusUnauthorized,
		"Put/RW/NoAuth/SlashInID.status":             http.StatusUnauthorized,
		"Put/RW/NoAuth/Conflict.status":              http.StatusUnauthorized,

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
		"Replicate.timeoutSeconds": 5,
		"Replicate.mode":           "pouchdb",

		"Query/RW/Admin/WithoutDocs/ScanDoc.status": http.StatusBadRequest,
		"Query/RW/NoAuth/WithoutDocs.status":        http.StatusUnauthorized,
		"Query/RW/NoAuth/WithDocs.status":           http.StatusUnauthorized,

		"ViewCleanup/RW/NoAuth.status": http.StatusUnauthorized,

		"Changes/Continuous.options": kivik.Params(map[string]interface{}{
			"live":    true,
			"timeout": false,
		}),
		"Changes/Continuous/RW/NoAuth.status": http.StatusUnauthorized,
		"Changes/Normal/RW/NoAuth.status":     http.StatusUnauthorized,
		"Changes/Continuous.skip": func() bool {
			// Node.js 14, required for GopherJS 1.17, does not support the AbortController function
			if ver := runtime.Version(); strings.HasPrefix(ver, "go1.17") {
				return true
			}
			return false
		}(),
		"Changes/Normal.skip": func() bool {
			// Node.js 14, required for GopherJS 1.17, does not support the AbortController function
			if ver := runtime.Version(); strings.HasPrefix(ver, "go1.17") {
				return true
			}
			return false
		}(),

		"DBsStats/NoAuth.status": http.StatusUnauthorized,

		"Copy/RW/NoAuth.status":                   http.StatusUnauthorized,
		"Compact/RW/NoAuth.status":                http.StatusUnauthorized,
		"Find/NoAuth.status":                      http.StatusUnauthorized,
		"Find/RW/NoAuth.status":                   http.StatusUnauthorized,
		"Explain/NoAuth.status":                   http.StatusUnauthorized,
		"CreateDoc/RW/NoAuth.status":              http.StatusUnauthorized,
		"Explain/RW/NoAuth.status":                http.StatusUnauthorized,
		"DeleteIndex/RW/NoAuth/ValidIndex.status": http.StatusUnauthorized,
	})
}

// PouchLocalTest runs the PouchDB tests against a local database.
func PouchLocalTest(t *testing.T) {
	client, err := kivik.New("pouch", "")
	if err != nil {
		t.Errorf("Failed to connect to PouchDB driver: %s", err)
		return
	}
	t.Cleanup(func() { _ = client.Close() })
	clients := &kt.Context{
		RW:    true,
		Admin: client,
	}
	kiviktest.RunTestsInternal(t, clients, kiviktest.SuitePouchLocal)
}

// PouchRemoteTest runs the PouchDB tests against a remote CouchDB database.
func PouchRemoteTest(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuitePouchRemote, "KIVIK_TEST_DSN_COUCH33")
}
