# SQLite Driver TODO

## Low Priority (polyfilled by kivik)

- [ ] **BulkDocs** (`db.go`) — kivik emulates via individual Put/CreateDoc.
- [ ] **Copy** (`db.go`) — kivik emulates via Get+Put.

## Performance / Code Quality

- [ ] **Find cross-type comparison correctness** (`find.go`) —
  `selectorToSQL` translates comparison operators (`$lt`, `$lte`, `$gt`,
  `$gte`, `$eq`, `$in`) to SQL, but SQLite doesn't support CouchDB's
  cross-type ordering (null < bool < number < string < array < object).
  `$in` also uses SQL `IN` which doesn't match CouchDB's deep equality
  for non-scalar or mixed-type values. When `selectorComplete=true`, the
  in-memory filter is skipped, producing incorrect results for these
  queries.
- [ ] **`use_index` doesn't influence query execution** (`find.go`) — The hint
  is validated and triggers a warning if missing, but doesn't guide the
  query plan.
- [ ] **Reduce caching** (`README.md`) — Reduce functions run on-demand with no
  intermediate result caching.
- [ ] **Mango SQL optimization** (`find.go`) — These selectors work via
  in-memory fallback but aren't translated to SQL: `$nor`, `$nin`,
  `$regex`, `$mod`, `$all`, `$elemMatch`, `$type`, `$size`, `$allMatch`,
  `$keyMapMatch`.
- [ ] **Filter in Go instead of SQL** (`query.go:569`) — Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.
