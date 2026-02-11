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

- [ ] **validate_doc_update not evaluated** (`put_test.go:1116`). Stored but
  never invoked.

- [ ] **Reduce caching** (`README.md`). Reduce functions run on-demand with no
  intermediate result caching.

## Code Quality

- [ ] **Ping placement** (`db.go:50`). TODO in code: "I think Ping belongs on
  \*client, not \*db". Requires v5 release (breaking API change).

- [ ] **Filter in Go instead of SQL** (`query.go:568`). Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.

## Integration Tests

See `test/INTEGRATION_TEST_PLAN.md` for the incremental plan to enable the
`kiviktest` integration suite.
