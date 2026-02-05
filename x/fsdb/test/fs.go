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

// Package test configures the integration test suite.
package test

import (
	"net/http"

	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func registerFSSuite() {
	kiviktest.RegisterSuite(kiviktest.SuiteKivikFS, kt.SuiteConfig{
		"AllDBs.expected": []string{},

		"CreateDB/RW/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": http.StatusPreconditionFailed,

		"AllDocs.skip": true, // FIXME: Not yet implemented
		// "AllDocs/Admin.databases":  []string{"foo"},
		// "AllDocs/Admin/foo.status": http.StatusNotFound,

		"DBExists/Admin.databases":      []string{"chicken"},
		"DBExists/Admin/chicken.exists": false,
		"DBExists/RW/Admin.exists":      true,

		"DestroyDB/RW/Admin/NonExistantDB.status": http.StatusNotFound,

		"Version.version":        `^0\.0\.1$`,
		"Version.vendor":         "Kivik",
		"Version.vendor_version": `^0\.0\.1$`,

		// Replications not to be implemented
		"GetReplications.skip": true,
		"Replicate.skip":       true,

		"Get/RW/Admin/bogus.status":  http.StatusNotFound,
		"Get/RW/NoAuth/bogus.status": http.StatusNotFound,

		"GetRev.skip":            true,                      // FIXME: Unimplemented
		"Flush.skip":             true,                      // FIXME: Unimplemented
		"Delete.skip":            true,                      // FIXME: Unimplemented
		"Stats.skip":             true,                      // FIXME: Unimplemented
		"CreateDoc.skip":         true,                      // FIXME: Unimplemented
		"Compact.skip":           true,                      // FIXME: Unimplemented
		"Security.skip":          true,                      // FIXME: Unimplemented
		"DBUpdates.status":       http.StatusNotImplemented, // FIXME: Unimplemented
		"Changes.skip":           true,                      // FIXME: Unimplemented
		"Copy.skip":              true,                      // FIXME: Unimplemented, depends on Get/Put or Copy
		"BulkDocs.skip":          true,                      // FIXME: Unimplemented
		"GetAttachment.skip":     true,                      // FIXME: Unimplemented
		"GetAttachmentMeta.skip": true,                      // FIXME: Unimplemented
		"PutAttachment.skip":     true,                      // FIXME: Unimplemented
		"DeleteAttachment.skip":  true,                      // FIXME: Unimplemented
		"Query.skip":             true,                      // FIXME: Unimplemented
		"Find.skip":              true,                      // FIXME: Unimplemented
		"Explain.skip":           true,                      // FIXME: Unimplemented
		"CreateIndex.skip":       true,                      // FIXME: Unimplemented
		"GetIndexes.skip":        true,                      // FIXME: Unimplemented
		"DeleteIndex.skip":       true,                      // FIXME: Unimplemented

		"Put/RW/Admin/LeadingUnderscoreInID.status":  http.StatusBadRequest,
		"Put/RW/Admin/Conflict.status":               http.StatusConflict,
		"Put/RW/NoAuth/LeadingUnderscoreInID.status": http.StatusBadRequest,
		"Put/RW/NoAuth/DesignDoc.status":             http.StatusUnauthorized,
		"Put/RW/NoAuth/Conflict.status":              http.StatusConflict,

		"SetSecurity.skip": true, // FIXME: Unimplemented
		"ViewCleanup.skip": true, // FIXME: Unimplemented
		"Rev.skip":         true, // FIXME: Unimplemented
	})
}
