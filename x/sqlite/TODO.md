# SQLite Driver TODO

## Missing Features

### Unimplemented methods (`db.go`)

These return a bare `"not implemented"` error:

- [ ] BulkDocs (low priority — kivik emulates via individual Put/CreateDoc)
- [ ] Copy (low priority — kivik emulates via Get+Put)
- [ ] Explain

### Unimplemented on `client` (`sqlite.go`)

- [ ] **DBUpdates** — Needs:
  - Database deletion events: Log "deleted" type in addition to "created"
  - Live notifications: Implement channel-based streaming for `feed=continuous`
  - See `dbupdates.go` and `createdb.go` for current implementation

### Incomplete features

- [ ] **Reduce caching** (`README.md`). Reduce functions run on-demand with no
  intermediate result caching.

## Code Quality

- [ ] **Filter in Go instead of SQL** (`query.go:568`). Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.

## Design Notes

### DBUpdates Architecture

**Current implementation (v0.1):**
1. ✅ **Persistence:** `kivik$db_updates_log` table with seq, db_name, type
2. ✅ **Sequence management:** Global auto-incrementing counter via AUTOINCREMENT
3. ✅ **Historical query:** `DBUpdates(since=N)` filters with `WHERE seq > N`
4. ⏳ **Live notification:** Stub implementation; channels not yet connected
5. ✅ **Multi-instance:** Persisted data queryable from any `Client` instance
6. ✅ **Feed types:** Validates feed parameter (normal, longpoll, continuous)

**Remaining for v0.2:**
- Log "deleted" events in addition to "created" events
- Implement channel-based live streaming for `feed=continuous`

**Constraints:**
- Single-process access assumed (SQLite limitation); multi-process access possible but not tested
- Testing-only use case (per README), so can be simpler than full CouchDB clustering
- DBUpdates is a replication feature, not just a test helper

## Integration Tests

See `test/INTEGRATION_TEST_PLAN.md` for the incremental plan to enable the
`kiviktest` integration suite.
