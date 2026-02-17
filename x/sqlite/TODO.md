# SQLite Driver TODO

## Missing Features

### Unimplemented methods (`db.go`)

These return a bare `"not implemented"` error:

- [ ] BulkDocs (low priority — kivik emulates via individual Put/CreateDoc)
- [ ] Copy (low priority — kivik emulates via Get+Put)
- [ ] Explain

### Incomplete features

- [ ] **Reduce caching** (`README.md`). Reduce functions run on-demand with no
  intermediate result caching.

## Code Quality

- [ ] **Filter in Go instead of SQL** (`query.go:568`). Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.

## Design Notes

### DBUpdates Redesign: Use `_global_changes` as backing store

Align with CouchDB spec: `_db_updates` reads from the `_global_changes` database
(a real, user-visible database), not a private internal table. Like CouchDB,
`_global_changes` is not created automatically — callers must run `ClusterSetup`
first (mirroring the CouchDB setup wizard).

**Key design decisions:**
- `_global_changes` is NOT auto-created at init (matching CouchDB pre-wizard state)
- `logGlobalChange` silently no-ops if `_global_changes` tables are absent
- `DBUpdates` returns HTTP 503 if `_global_changes` is absent
- Each DB event → one new append-only document: `{"db_name": "...", "type": "created|updated|deleted"}`
- `DBUpdates` wraps `db.Changes(include_docs=true)` on `_global_changes`

**Implementation cycles (TDD):**

- [ ] **Cycle 1:** Implement `driver.Cluster`: `ClusterSetup` creates `_users`/`_replicator`/`_global_changes` for `"enable_single_node"`; `ClusterStatus` returns `"cluster_disabled"`/`"single_node_enabled"`; `Membership` returns single fake node
- [ ] **Cycle 2:** DB creation logs to `_global_changes` (silently no-ops if absent)
- [ ] **Cycle 3:** DB deletion logs to `_global_changes` (silently no-ops if absent)
- [ ] **Cycle 4:** `DBUpdates` reads `_global_changes/_changes`; returns 503 if absent
- [ ] **Cycle 5:** Remove `kivik$db_updates_log` (table, log function, template func, old structs)

**Critical files:**
- `dbupdates.go` — rewrite `DBUpdates`; add `globalChangesDBUpdates` adapter; add `logGlobalChange`
- `createdb.go` — replace `logDBUpdate` with `logGlobalChange`
- `sqlite.go` — replace `logDBUpdate` with `logGlobalChange` in `DestroyDB`; remove `ensureDBUpdatesLog`
- `schema.go` — remove `kivik$db_updates_log` from schema
- `templ.go` — remove `DBUpdatesLog()` template func

**Constraints:**
- Single-process access assumed (SQLite limitation)
- `_global_changes` sequences are incompatible with old `kivik$db_updates_log` sequences (acceptable)

## Integration Tests

See `test/INTEGRATION_TEST_PLAN.md` for the incremental plan to enable the
`kiviktest` integration suite.
