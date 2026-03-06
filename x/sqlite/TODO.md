# SQLite Driver TODO

## Correctness

- [ ] **View index invalidation on ddoc update** (`designdocs.go`) — Verify
  that old view indexes are properly dropped/rebuilt when a design document
  changes.

## Functionality

- [ ] **`_approx_count_distinct` reduce function** — Not implemented.
- [ ] **`_sum` extended capabilities** — CouchDB supports summing arrays and
  objects; verify full compatibility.
  See: https://docs.couchdb.org/en/stable/ddocs/ddocs.html#sum
- [ ] **`_stats` edge cases** — Differing lengths of arrays of floats, arrays
  with mixed types, arrays of stats objects.
- [ ] **Map/reduce timeout** — No timeout handling for long-running JS
  functions.
- [ ] **`revs=true` + attachments in OpenRevs** — Not implemented/tested.
- [ ] **Historical revision attachments** — Get attachment from old revision,
  correct attachment in conflict scenarios.
- [ ] **Offset/TotalRows on AllDocs rows** — Verify these return correct values.
- [ ] **AllDocs on nonexistent DB** — Verify correct error behavior.
- [ ] **Filter functions** — Need more comprehensive testing/fleshing out.

## Test Gaps

- [ ] **Changes feed `conflicts` option** — Not tested.
- [ ] **Changes feed mode coverage** — Longpoll missing tests for descending,
  filter, doc_ids, style. Continuous has minimal test coverage.
- [ ] **DeleteAttachment on missing DB** — Should return "db not found".
- [ ] **CreateDoc edge cases** — nil doc, UUID configuration options,
  duplicate UUID retry.
- [ ] **Put with update function interaction** — Not tested.
- [ ] **GetAttachment edge cases** — Attachment from historical revision,
  correct attachment in conflict, various 404 scenarios.
- [ ] **Purge edge cases** — Purging leaf + parent simultaneously.
- [ ] **Design doc edge cases** — Unsupported language handling, func_type
  update/validate storage.

## Low Priority (polyfilled by kivik)

- [ ] **BulkDocs** (`db.go`) — kivik emulates via individual Put/CreateDoc.
- [ ] **Copy** (`db.go`) — kivik emulates via Get+Put.

## Performance / Code Quality

- [ ] **Reduce caching** (`README.md`) — Reduce functions run on-demand with no
  intermediate result caching.
- [ ] **Filter in Go instead of SQL** (`query.go:569`) — Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.
- [ ] **Cross-type SQL optimization** (`find.go`) — Inequality SQL type guards
  could encode CouchDB's type ordering directly (e.g. `$gt: 21` →
  `(type IN ('integer','real') AND val > 21) OR type IN ('text','array','object')`)
  to allow `selectorComplete=true` and skip the in-memory fallback. Currently
  correctness is ensured by always falling back to in-memory filtering.
