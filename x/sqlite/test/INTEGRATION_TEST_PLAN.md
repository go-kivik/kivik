# Plan: Enable SQLite Driver Integration Tests

## Background

The SQLite driver (`x/sqlite/`) has zero integration test coverage via the shared
`kiviktest` framework. Every test in `x/sqlite/test/test.go` is set to `.skip: true`.
The driver has unit tests, but the `kiviktest` suite validates CouchDB API compatibility
across all drivers.

## Prerequisite: Fix AllDBs prefix bug

**File:** `x/sqlite/alldbs.go`

`AllDBs()` returns raw SQLite table names (e.g., `kivik$mydb`), but all other methods
(`CreateDB`, `DestroyDB`, `DBExists`, `DB`) expect the logical name without the prefix
(e.g., `mydb`). This breaks the `kiviktest` cleanup routine, which calls
`AllDBs()` then passes results to `DestroyDB()`.

Fix: strip `tablePrefix` from each returned name.

This is a separate commit before test changes begin.

## Integration Test Enablement

**File to modify:** `x/sqlite/test/test.go`

Enable one test suite at a time. For each, remove the `.skip` entry and add the
required config keys. Run the tests, fix config if needed.

Tests that remain permanently skipped (not implemented by the driver):
- Compact, CreateIndex, DBUpdates, DeleteIndex, Explain, Flush, GetIndexes
- GetReplications, Security, SetSecurity, ViewCleanup
- Stats, Copy, Replicate, BulkDocs, DBsStats, AllDBsStats, Session

### Task 1: Version

Remove `"Version.skip": true`. Add:
```go
"Version.version": `^0\.0\.1$`,
"Version.vendor":  `^Kivik$`,
```

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/Version -v`

### Task 2: CreateDB + DestroyDB

Remove both `.skip` entries. Add:
```go
"CreateDB/RW/Admin/Recreate.status":        http.StatusPreconditionFailed,
"DestroyDB/RW/Admin/NonExistantDB.status":  http.StatusNotFound,
```

Verify: `go test ./x/sqlite/test/ -run 'TestSQLite/kivikSQLite/(CreateDB|DestroyDB)' -v`

### Task 3: DBExists + AllDBs

Remove both `.skip` entries. Add:
```go
"AllDBs.expected":                   []string{},
"DBExists/Admin.databases":          []string{"chicken"},
"DBExists/Admin/chicken.exists":     false,
"DBExists/RW/group/Admin.exists":    true,
```

Verify: `go test ./x/sqlite/test/ -run 'TestSQLite/kivikSQLite/(DBExists|AllDBs)' -v`

### Task 4: Put

Remove `"Put.skip": true`. Add:
```go
"Put/RW/Admin/group/LeadingUnderscoreInID.status": http.StatusBadRequest,
"Put/RW/Admin/group/Conflict.status":              http.StatusConflict,
```

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/Put -v`

### Task 5: Get + GetRev

Remove both `.skip` entries. Add:
```go
"Get/RW/group/Admin/bogus.status":    http.StatusNotFound,
"GetRev/RW/group/Admin/bogus.status": http.StatusNotFound,
```

Verify: `go test ./x/sqlite/test/ -run 'TestSQLite/kivikSQLite/(Get|GetRev)/' -v`

### Task 6: Delete

Remove `"Delete.skip": true`. Add:
```go
"Delete/RW/Admin/group/MissingDoc.status":       http.StatusNotFound,
"Delete/RW/Admin/group/InvalidRevFormat.status":  http.StatusBadRequest,
"Delete/RW/Admin/group/WrongRev.status":          http.StatusConflict,
```

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/Delete -v`

### Task 7: CreateDoc

Remove `"CreateDoc.skip": true`. No additional config needed.

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/CreateDoc -v`

### Task 8: AllDocs

Remove `"AllDocs.skip": true`. Add:
```go
"AllDocs.databases": []string{},
```

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/AllDocs -v`

### Task 9: Find

Remove `"Find.skip": true`. Add:
```go
"Find.databases":                      []string{},
"Find/RW/group/Admin/Warning.warning": "",
```

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/Find -v`

### Task 10: Query

Remove `"Query.skip": true`. Add:
```go
"Query/RW/group/Admin/WithoutDocs/ScanDoc.status": http.StatusBadRequest,
```

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/Query -v`

### Task 11: PutAttachment

Remove `"PutAttachment.skip": true`. Add:
```go
"PutAttachment/RW/group/Admin/Conflict.status": http.StatusConflict,
```

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/PutAttachment -v`

### Task 12: GetAttachment + GetAttachmentMeta

Remove both `.skip` entries. Add:
```go
"GetAttachment/RW/group/Admin/foo/NotFound.status":     http.StatusNotFound,
"GetAttachmentMeta/RW/group/Admin/foo/NotFound.status":  http.StatusNotFound,
```

Verify: `go test ./x/sqlite/test/ -run 'TestSQLite/kivikSQLite/(GetAttachment|GetAttachmentMeta)/' -v`

### Task 13: DeleteAttachment

Remove `"DeleteAttachment.skip": true`. Add:
```go
"DeleteAttachment/RW/group/Admin/NotFound.status": http.StatusNotFound,
"DeleteAttachment/RW/group/Admin/NoDoc.status":    http.StatusConflict,
```

Note: `NoDoc.status` may need adjustment - the driver might return `StatusNotFound`
instead of `StatusConflict`. Verify and adjust.

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/DeleteAttachment -v`

### Task 14: Changes

Remove `"Changes.skip": true`. Add:
```go
"Changes/Continuous.options": kivik.Params(map[string]interface{}{
    "feed":      "longpoll",
    "since":     "now",
    "heartbeat": 6000,
}),
```

The SQLite driver supports `longpoll` but not `continuous` feed mode.

Verify: `go test ./x/sqlite/test/ -run TestSQLite/kivikSQLite/Changes -v`

## Notes

- The driver is single-user (no auth model). The test runner only sets `Admin`,
  not `NoAuth`, so all `RunNoAuth` sub-tests auto-skip.
- Config values for error status codes are based on the MemoryDB reference config
  and the SQLite driver's error handling code. Some may need adjustment during
  implementation.
- If any test reveals a driver bug, leave a TODO comment in the test config and
  keep the test skipped, per project conventions.

## Verification

After all tasks, run the full integration test suite:
```
go test ./x/sqlite/test/ -v 2>&1 | tee /tmp/sqlite-integration.out
```
