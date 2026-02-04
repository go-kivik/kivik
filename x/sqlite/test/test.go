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

// Package test manages the integration tests for the SQLite driver.
package test

import (
	"net/http"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

// RegisterSQLiteSuites registers the SQLite related integration test suites.
func RegisterSQLiteSuites() {
	registerSuiteSQLite()
}

func registerSuiteSQLite() {
	kiviktest.RegisterSuite(kiviktest.SuiteKivikSQLite, kt.SuiteConfig{
		// Skip all tests that are not applicable to SQLite
		"Compact.skip":         true,
		"CreateIndex.skip":     true,
		"DBUpdates.skip":       true,
		"DeleteIndex.skip":     true,
		"Explain.skip":         true,
		"Flush.skip":           true,
		"GetIndexes.skip":      true,
		"GetReplications.skip": true,
		"Security.skip":        true,
		"SetSecurity.skip":     true,
		"ViewCleanup.skip":     true,

		"Version.version": `^0\.0\.1$`,
		"Version.vendor":  `^Kivik$`,
		"Put/RW/Admin/group/LeadingUnderscoreInID.status":      http.StatusBadRequest,
		"Put/RW/Admin/group/Conflict.status":                   http.StatusConflict,
		"Get/RW/group/Admin/bogus.status":                      http.StatusNotFound,
		"Delete/RW/Admin/group/MissingDoc.status":              http.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status":        http.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":                http.StatusConflict,
		"DestroyDB/RW/Admin/NonExistantDB.status":              http.StatusNotFound,
		"AllDocs.databases":                                    []string{},
		"Stats.skip":                                           true,
		"Copy.skip":                                            true,
		"GetAttachment/RW/group/Admin/foo/NotFound.status":     http.StatusNotFound,
		"PutAttachment/RW/group/Admin/Conflict.status":         http.StatusConflict,
		"DeleteAttachment/RW/group/Admin/NotFound.status":      http.StatusNotFound,
		"DeleteAttachment/RW/group/Admin/NoDoc.status":         http.StatusNotFound,
		"Replicate.skip":                                       true,
		"Find.databases":                                       []string{},
		"Find/RW/group/Admin/Warning.warning":                  "",
		"GetAttachmentMeta/RW/group/Admin/foo/NotFound.status": http.StatusNotFound,
		"DBExists/Admin.databases":                             []string{"chicken"},
		"DBExists/Admin/chicken.exists":                        false,
		"DBExists/RW/group/Admin.exists":                       true,
		"Query/RW/group/Admin/WithoutDocs/ScanDoc.status":      http.StatusBadRequest,
		"Changes/Continuous.options": kivik.Params(map[string]interface{}{
			"feed":  "continuous",
			"since": "now",
		}),
		"CreateDB/RW/Admin/Recreate.status":  http.StatusPreconditionFailed,
		"GetRev/RW/group/Admin/bogus.status": http.StatusNotFound,
		"BulkDocs.skip":                      true,
		"DBsStats.skip":                      true,
		"AllDBsStats.skip":                   true,
		"Session.skip":                       true,
	})
}
