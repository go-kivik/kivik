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

		"Version.skip":           true,
		"Put.skip":               true,
		"Get.skip":               true,
		"Delete.skip":            true,
		"CreateDoc.skip":         true,
		"DestroyDB.skip":         true,
		"AllDocs.skip":           true,
		"AllDBs.skip":            true,
		"Stats.skip":             true,
		"Copy.skip":              true,
		"GetAttachment.skip":     true,
		"PutAttachment.skip":     true,
		"DeleteAttachment.skip":  true,
		"Replicate.skip":         true,
		"Find.skip":              true,
		"GetAttachmentMeta.skip": true,
		"DBExists.skip":          true,
		"Query.skip":             true,
		"Changes.skip":           true,
		"CreateDB.skip":          true,
		"GetRev.skip":            true,
		"BulkDocs.skip":          true,
		"DBsStats.skip":          true,
		"AllDBsStats.skip":       true,
		"Session.skip":           true,
	})
}
