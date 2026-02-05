# SQLite Driver TODO

## Missing Features

### Unimplemented methods (`db.go`)

These return a bare `"not implemented"` error:

- [ ] BulkDocs (low priority — kivik emulates via individual Put/CreateDoc)
- [ ] Copy (low priority — kivik emulates via Get+Put)
- [ ] CreateIndex
- [ ] DeleteIndex
- [ ] Explain
- [ ] GetIndexes
- [ ] Stats

### Unimplemented on `client` (`sqlite.go`)

- [ ] Replicate / GetReplications
- [ ] DBUpdates
- [ ] Session
- [ ] Security / SetSecurity

### Incomplete features

- [ ] **Find: sort** (`find_test.go:193`). Sort returns an error; not
  implemented. Other unimplemented Find options noted at `find_test.go:219-227`:
  stable, update, stale, use_index, execution_stats.

- [ ] **validate_doc_update not evaluated** (`designdocs.go:67-68`). The
  function body is stored when a design document is written, but never called
  during Put or CreateDoc. See also `put_test.go:1115`.

- [ ] **Update functions not evaluated** (`put_test.go:1116`). Stored but never
  invoked.

- [ ] **RevsDiff: possible_ancestors** (`revsdiff_test.go:59`). The response
  never populates `possible_ancestors`.

- [ ] **Attachment compression** (`json.go:244`). Encoding and encoded_length
  fields are stubbed out.

- [ ] **Reduce caching** (`README.md`). Reduce functions run on-demand with no
  intermediate result caching.

## Code Quality

- [ ] **Ping placement** (`db.go:50`). TODO in code: "I think Ping belongs on
  \*client, not \*db".

- [ ] **Filter in Go instead of SQL** (`query.go:568`). Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.

## Integration Tests

See `test/INTEGRATION_TEST_PLAN.md` for the incremental plan to enable the
`kiviktest` integration suite.
