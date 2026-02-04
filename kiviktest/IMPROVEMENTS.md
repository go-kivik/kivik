# kiviktest: Gradual Improvement Roadmap

Constraint: minimum Go version is 1.19 (`go.mod`).

---

## Batch 1: Immediate cleanup (each a separate commit, zero behavior change)

### 1a. Delete dead code in kt/kt.go
- Delete commented-out retry conditions at lines 364-368

### 1b. Remove `fmt.Printf` in Retry
- `kt/kt.go:326`: `fmt.Printf("Retrying after error: %s\n", err)` bypasses test output capture
- Delete the line. The error is already wrapped and returned to the caller.

### 1c. Delete `go110testing.go`
- Build constraint `go1.10 && !go1.18` never compiles with Go 1.19+ minimum
- Pure dead code

### 1d. Simplify `go118testing.go` build constraint
- Change `//go:build go1.18 && !go1.23` to `//go:build !go1.23`

### 1e. `interface{}` -> `any` across the package
- Mechanical replacement in all files under `kiviktest/`
- Includes `kt/config.go` (`SuiteConfig map[string]interface{}`), `kt/kt.go`, shim files, and test files

### Verification:
- `golangci-lint run ./kiviktest/...`
- `go build ./kiviktest/...`
- Run CouchDB integration tests if available (`USETC=1 go test ./couchdb/test/...`)

---

## Batch 2: Add `ctx.DB()` helper (boilerplate reduction)

Add helper methods to `kt.Context`:

```go
func (c *Context) DB(client *kivik.Client, dbname string) *kivik.DB
func (c *Context) AdminDB(dbname string) *kivik.DB
```

Then update call sites across `client/` and `db/` test files.

---

## Batch 3: Remove "group" subtest wrappers

Since `TestDB()` already uses `t.Cleanup()` (kt.go:248), the "group" subtest wrapper is unnecessary. It was originally needed so `defer` wouldn't run before parallel subtests completed. With `t.Cleanup`, the testing framework handles this.

Steps:
1. Remove `ctx.Run("group", ...)` wrappers from all test files, promoting their contents one level up
2. Update all suite config keys that contain `/group/` (e.g., `"Put/RW/Admin/group/Create.status"` -> `"Put/RW/Admin/Create.status"`)

---

## Batch 4: Shared CouchDB config base

Extract common config entries into a base map, with per-version overrides.

Add a `SuiteConfig.Merge(other SuiteConfig)` or just use Go maps merge.

---

## Batch 5: Fix bulk.go complexity

Extract repeated BulkResult checking into a helper in `db/bulk.go`.

Remove `// nolint: gocyclo` after complexity is reduced.

---

## Longer-term: Evaluate replacing kt.Context

`kt.Context` wraps `*testing.T` to provide:
1. Admin/NoAuth client pair management
2. Config-driven error expectations (`CheckError`/`IsExpectedSuccess`)
3. Test DB lifecycle helpers (`TestDB`, `DestroyDB`)
4. Thin delegation (`Errorf`, `Fatalf`, `Parallel`, etc.)

Item 4 adds no value. Items 1 and 3 could be a plain struct without wrapping `*testing.T`. Item 2 is the unique value proposition.

A possible replacement architecture:
- Tests use `*testing.T` directly
- A `SuiteHelper` struct holds clients + config (not wrapping T)
- Error expectation checking becomes `suite.ExpectSuccess(t, err)` instead of `ctx.IsExpectedSuccess(err)`
- `TestDB(t)` becomes a method on SuiteHelper

This would be a file-by-file migration, not a big-bang rewrite.
