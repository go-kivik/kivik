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
	"context"
	"net/http"
	"testing"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

// RegisterMemoryDBSuite registers the MemoryDB integration test suite.
func RegisterMemoryDBSuite() {
	kiviktest.RegisterSuite(kiviktest.SuiteKivikMemory, kt.SuiteConfig{
		// Unsupported features
		"Flush.skip": true,

		"AllDBs.expected": []string{"_users"},

		"CreateDB/RW/NoAuth.status":         http.StatusUnauthorized,
		"CreateDB/RW/Admin/Recreate.status": http.StatusPreconditionFailed,

		"AllDocs.skip": true, // FIXME: Unimplemented

		"DBExists/Admin.databases":      []string{"chicken"},
		"DBExists/Admin/chicken.exists": false,
		"DBExists/RW/Admin.exists":      true,

		"DestroyDB/RW/Admin/NonExistantDB.status": http.StatusNotFound,

		"Log.status":          http.StatusNotImplemented,
		"Log/Admin/HTTP.skip": true,

		"Version.version":        `^0\.0\.1$`,
		"Version.vendor":         `^Kivik Memory Adaptor$`,
		"Version.vendor_version": `^0\.0\.1$`,

		// Replications not to be implemented
		"GetReplications.skip": true,
		"Replicate.skip":       true,

		"Get/RW/Admin/bogus.status": http.StatusNotFound,

		"GetRev/RW/Admin/bogus.status": http.StatusNotFound,

		"Put/RW/Admin/LeadingUnderscoreInID.status": http.StatusBadRequest,
		"Put/RW/Admin/Conflict.status":              http.StatusConflict,

		"Delete/RW/Admin/MissingDoc.status":       http.StatusNotFound,
		"Delete/RW/Admin/InvalidRevFormat.status": http.StatusBadRequest,
		"Delete/RW/Admin/WrongRev.status":         http.StatusConflict,

		"Security.databases":            []string{"_users", "chicken", "_duck"},
		"Security/Admin/chicken.status": http.StatusNotFound,
		"Security/Admin/_duck.status":   http.StatusNotFound,

		"SetSecurity/RW/Admin/NotExists.status": http.StatusNotFound,

		"BulkDocs/RW/NoAuth/Mix/Conflict.status": http.StatusConflict,
		"BulkDocs/RW/Admin/Mix/Conflict.status":  http.StatusConflict,

		"Find.databases":                 []string{"chicken", "_duck"},
		"Find/Admin/chicken.status":      http.StatusNotFound,
		"Find/Admin/_duck.status":        http.StatusNotFound,
		"Find/NoAuth/chicken.status":     http.StatusNotFound,
		"Find/NoAuth/_duck.status":       http.StatusUnauthorized,
		"Find/RW/Admin/Warning.warning":  "no matching index found, create an index to optimize query time",
		"Find/RW/NoAuth/Warning.warning": "no matching index found, create an index to optimize query time",

		"Explain.skip":           true,                      // FIXME: Unimplemented
		"Stats.skip":             true,                      // FIXME: Unimplemented
		"Compact.skip":           true,                      // FIXME: Unimplemented
		"DBUpdates.status":       http.StatusNotImplemented, // FIXME: Unimplemented
		"Changes.skip":           true,                      // FIXME: Unimplemented
		"Copy.skip":              true,                      // FIXME: Unimplemented, depends on Get/Put or Copy
		"GetAttachment.skip":     true,                      // FIXME: Unimplemented
		"GetAttachmentMeta.skip": true,                      // FIXME: Unimplemented
		"PutAttachment.skip":     true,                      // FIXME: Unimplemented
		"DeleteAttachment.skip":  true,                      // FIXME: Unimplemented
		"Query.skip":             true,                      // FIXME: Unimplemented
		"CreateIndex.skip":       true,                      // FIXME: Unimplemented
		"GetIndexes.skip":        true,                      // FIXME: Unimplemented
		"DeleteIndex.skip":       true,                      // FIXME: Unimplemented
		"SetSecurity.skip":       true,                      // FIXME: Unimplemented
		"ViewCleanup.skip":       true,                      // FIXME: Unimplemented
	})
}

// MemoryTest runs the integration tests for the Memory driver.
func MemoryTest(t *testing.T) {
	t.Helper()
	client, err := kivik.New("memory", "")
	if err != nil {
		t.Fatalf("Failed to connect to memory driver: %s\n", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	clients := &kt.Context{
		RW:    true,
		Admin: client,
		T:     t,
	}
	if err := client.CreateDB(context.Background(), "_users"); err != nil {
		t.Fatal(err)
	}
	kiviktest.RunTestsInternal(clients, kiviktest.SuiteKivikMemory)
}
