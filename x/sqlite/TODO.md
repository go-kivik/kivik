# SQLite Driver TODO

## Low Priority (polyfilled by kivik)

- [ ] **BulkDocs** (`db.go`) — kivik emulates via individual Put/CreateDoc.
- [ ] **Copy** (`db.go`) — kivik emulates via Get+Put.

## Performance / Code Quality

- [ ] **Find cross-type comparison correctness** (`find.go`) —
  `selectorToSQL` translates comparison operators (`$lt`, `$lte`, `$gt`,
  `$gte`, `$eq`) to SQL, but SQLite doesn't support CouchDB's cross-type
  ordering (null < bool < number < string < array < object). When
  `selectorComplete=true`, the in-memory filter is skipped, producing
  incorrect results for cross-type queries.
- [ ] **`use_index` doesn't influence query execution** (`find.go`) — The hint
  is validated and triggers a warning if missing, but doesn't guide the
  query plan.
- [ ] **Reduce caching** (`README.md`) — Reduce functions run on-demand with no
  intermediate result caching.
- [ ] **Mango SQL optimization** (`find.go`) — These selectors could be
  translated to SQL for index support but aren't yet: `$size`, `$type`.
  The remaining operators (`$nin`, `$mod`, `$all`, `$elemMatch`,
  `$allMatch`, `$keyMapMatch`) aren't indexable in SQLite and are handled
  adequately by the in-memory fallback.
- [ ] **Filter in Go instead of SQL** (`query.go:569`) — Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.
