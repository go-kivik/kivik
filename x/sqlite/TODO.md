# SQLite Driver TODO

## Missing Features

### Unimplemented methods (`db.go`)

These return a bare `"not implemented"` error:

- [ ] BulkDocs (low priority — kivik emulates via individual Put/CreateDoc)
- [ ] Copy (low priority — kivik emulates via Get+Put)
- [ ] Explain

### Unimplemented on `client` (`sqlite.go`)

- [ ] **DBUpdates** — Partially implemented (live notifications only). Needs:
  - Persistence layer: `_local/db_updates_log` table to store `{seq, db_name, type}`
  - Historical replay: Query log for events since given sequence before subscribing to live updates
  - Multi-instance support: Currently only notifies channels within the same `Client` instance; needs to be queryable from any instance opening the database
  - The feature is fundamental to replication, not just testing. CouchDB's `_db_updates` supports `since` parameter and both `feed=normal` (historical) and `feed=continuous` (live streaming)
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

**Current approach (v0.1):**
- In-memory channels for live notifications within a single `Client` instance
- `notify()` method called after `CreateDB`/`DestroyDB` commits
- No persistence; no historical replay capability

**Needed for production replication (v0.2+):**
1. **Persistence:** Add `_local/db_updates_log` table with schema:
   ```
   seq INT PRIMARY KEY,
   db_name TEXT,
   type TEXT ('created'|'deleted'),
   created_at TIMESTAMP
   ```

2. **Sequence management:** Global auto-incrementing counter (already have `client.seq`)

3. **Historical query:** `DBUpdates(since=N)` should:
   - SELECT from log WHERE seq > N
   - Replay those events first
   - Then subscribe to live notifications

4. **Live notification:** Keep channel mechanism for within-process subscribers

5. **Multi-instance:** Any `Client` instance can call `DBUpdates` and read from the persisted log

6. **Feed types:** Support `feed=normal` (return all historical, then close) and `feed=continuous` (stream live)

**Constraints:**
- Single-process access assumed (SQLite limitation); multi-process access possible but not tested
- Testing-only use case (per README), so can be simpler than full CouchDB clustering
- DBUpdates is a replication feature, not just a test helper

## Integration Tests

See `test/INTEGRATION_TEST_PLAN.md` for the incremental plan to enable the
`kiviktest` integration suite.
