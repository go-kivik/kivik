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

// Package test manages the integration tests for the PostgreSQL driver.
package test

import (
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

// RegisterPGSuites registers the PostgreSQL related integration test suites.
func RegisterPGSuites() {
	registerSuitePG()
}

func registerSuitePG() {
	kiviktest.RegisterSuite(kiviktest.SuitePG, kt.SuiteConfig{
		"Version.version":        `^0\.0\.`,
		"Version.vendor":         `^Kivik$`,
		"Version.vendor_version": ``, // CouchDB 2.0+ no longer has a vendor version

		"Security.skip":          true,
		"DestroyDB.skip":         true,
		"Delete.skip":            true,
		"ViewCleanup.skip":       true,
		"Put.skip":               true,
		"AllDocs.skip":           true,
		"Explain.skip":           true,
		"AllDBs.skip":            true,
		"Stats.skip":             true,
		"Compact.skip":           true,
		"Copy.skip":              true,
		"GetAttachment.skip":     true,
		"PutAttachment.skip":     true,
		"DeleteAttachment.skip":  true,
		"Replicate.skip":         true,
		"Find.skip":              true,
		"CreateIndex.skip":       true,
		"GetAttachmentMeta.skip": true,
		"CreateDoc.skip":         true,
		"Flush.skip":             true,
		"DBExists.skip":          true,
		"Query.skip":             true,
		"Get.skip":               true,
		"Changes.skip":           true,
		"CreateDB.skip":          true,
		"GetRev.skip":            true,
		"BulkDocs.skip":          true,
		"DBsStats.skip":          true,
		"AllDBsStats.skip":       true,
		"Session.skip":           true,
		"GetIndexes.skip":        true,
		"DeleteIndex.skip":       true,
		"SetSecurity.skip":       true,
		"DBUpdates.skip":         true,
		"GetReplications.skip":   true,
	})
}
