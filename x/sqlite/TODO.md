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

## Bugs

- [ ] **Single quotes in JSON field names break queries**. `selectorToSQL()`,
  `fieldCondition()`, `inequalityCondition()` (`find.go`), and `CreateIndex()`
  (`indexes.go`) embed `mango.FieldToJSONPath()` output directly into
  single-quoted SQL string literals. A field name containing `'` (e.g.
  `foo'bar`) causes a syntax error. Not SQL injection — the query fails before
  executing — but prevents legitimate use of such field names. Fix by switching
  to double-quoted path strings or parameterizing the path argument.

- [ ] **Potential panic on empty json.RawMessage** (`find.go`).
  `fieldCondition()` accesses `val[0]` without a length check. An empty
  `json.RawMessage` would panic. Unlikely in practice since `json.Unmarshal`
  won't produce one, but a defensive check is cheap.

## Code Quality

- [ ] **Ping placement** (`db.go:50`). TODO in code: "I think Ping belongs on
  \*client, not \*db". Requires v5 release (breaking API change).

- [ ] **Filter in Go instead of SQL** (`query.go:568`). Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.

## Integration Tests

See `test/INTEGRATION_TEST_PLAN.md` for the incremental plan to enable the
`kiviktest` integration suite.
