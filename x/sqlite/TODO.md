# SQLite Driver TODO

## Correctness

- [ ] **Find cross-type comparison correctness** (`find.go`) —
  `selectorToSQL` translates comparison operators (`$lt`, `$lte`, `$gt`,
  `$gte`, `$eq`) to SQL, but SQLite doesn't support CouchDB's cross-type
  ordering (null < bool < number < string < array < object). When
  `selectorComplete=true`, the in-memory filter is skipped, producing
  incorrect results for cross-type queries.
- [ ] **Purge sequence tracking** (`purge.go`, `schema.go`) — No purge_seq is
  stored or returned. `PurgeResult.Seq` always returns 0. Needs schema
  addition and tracking on each purge operation.
- [ ] **View index invalidation on ddoc update** (`designdocs.go`) — Verify
  that old view indexes are properly dropped/rebuilt when a design document
  changes.

## Functionality

- [ ] **`use_index` doesn't influence query execution** (`find.go`) — The hint
  is validated and triggers a warning if missing, but doesn't guide the
  query plan.
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
