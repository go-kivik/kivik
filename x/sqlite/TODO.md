# SQLite Driver TODO

## Missing Features

### Unimplemented methods (`db.go`)

These return a bare `"not implemented"` error:

- [ ] BulkDocs (low priority — kivik emulates via individual Put/CreateDoc)
- [ ] Copy (low priority — kivik emulates via Get+Put)
- [ ] Explain

### Unimplemented on `client` (`sqlite.go`)

- [ ] Replicate / GetReplications
- [ ] DBUpdates

### Incomplete features

- [ ] **Mango indexes are metadata-only**. CreateIndex/DeleteIndex/GetIndexes
  store and retrieve index definitions but do not create actual SQLite indexes
  or influence `_find` query planning. Making them functional requires creating
  expression indexes (e.g. on `json_extract(doc, '$.field')`) and updating the
  Find query path to use them.

- [ ] **Find: sort** (`find_test.go:193`). Sort returns an error; not
  implemented. Other unimplemented Find options noted at `find_test.go:219-227`:
  stable, update, stale, use_index, execution_stats.

- [ ] **Update functions not evaluated** (`put_test.go:1116`). Stored but never
  invoked.

- [ ] **Reduce caching** (`README.md`). Reduce functions run on-demand with no
  intermediate result caching.

### Ignored or missing options

Many functions accept `driver.Options` but ignore some or all of them.

Note: `batch=ok` is intentionally not implemented for Put, Delete, and CreateDoc.
It's a CouchDB durability optimization that doesn't apply to SQLite.

- [ ] **Find** (`find.go:21`). Options `update`, `stale`, and `use_index` are
  no-ops until Mango indexes are functional (not just metadata-only).
  `stable` is permanently a no-op (single-node SQLite has no shards).

## Code Quality

- [ ] **Ping placement** (`db.go:50`). TODO in code: "I think Ping belongs on
  \*client, not \*db". Requires v5 release (breaking API change).

- [ ] **Filter in Go instead of SQL** (`query.go:568`). Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.


## Integration Tests

See `test/INTEGRATION_TEST_PLAN.md` for the incremental plan to enable the
`kiviktest` integration suite.
