# kiviktest: Gradual Improvement Roadmap

Constraint: minimum Go version is 1.20 (`go.mod`).

---

## Batch 2: Add `ctx.DB()` helper (boilerplate reduction)

Add helper methods to `kt.Context`:

```go
func (c *Context) DB(client *kivik.Client, dbname string) *kivik.DB
func (c *Context) AdminDB(dbname string) *kivik.DB
```

Then update call sites across `client/` and `db/` test files.

---

## Batch 3: Shared CouchDB config base

Extract common config entries into a base map, with per-version overrides.

Add a `SuiteConfig.Merge(other SuiteConfig)` or just use Go maps merge.

---

## Batch 4: Fix bulk.go complexity

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
