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
		// TODO: The driver should reject document IDs with leading underscores (except _design/ and _local/).
		"Put/RW/Admin/group/LeadingUnderscoreInID.skip": true,
		"Put/RW/Admin/group/Conflict.status":            http.StatusConflict,
		"Get/RW/group/Admin/bogus.status":               http.StatusNotFound,
		"Delete/RW/Admin/group/MissingDoc.status":       http.StatusNotFound,
		"Delete/RW/Admin/group/InvalidRevFormat.status": http.StatusBadRequest,
		"Delete/RW/Admin/group/WrongRev.status":         http.StatusConflict,
		"CreateDoc.skip":                                true,
		"DestroyDB/RW/Admin/NonExistantDB.status":       http.StatusNotFound,
		"AllDocs.skip":                       true,
		"AllDBs.expected":                    []string{},
		"Stats.skip":                         true,
		"Copy.skip":                          true,
		"GetAttachment.skip":                 true,
		"PutAttachment.skip":                 true,
		"DeleteAttachment.skip":              true,
		"Replicate.skip":                     true,
		"Find.skip":                          true,
		"GetAttachmentMeta.skip":             true,
		"DBExists/Admin.databases":           []string{"chicken"},
		"DBExists/Admin/chicken.exists":      false,
		"DBExists/RW/group/Admin.exists":     true,
		"Query.skip":                         true,
		"Changes.skip":                       true,
		"CreateDB/RW/Admin/Recreate.status":  http.StatusPreconditionFailed,
		"GetRev/RW/group/Admin/bogus.status": http.StatusNotFound,
		"BulkDocs.skip":                      true,
		"DBsStats.skip":                      true,
		"AllDBsStats.skip":                   true,
		"Session.skip":                       true,
	})
}
