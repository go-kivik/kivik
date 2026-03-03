# SQLite Driver TODO

## Low Priority (polyfilled by kivik)

- [ ] **BulkDocs** (`db.go`) — kivik emulates via individual Put/CreateDoc.
- [ ] **Copy** (`db.go`) — kivik emulates via Get+Put.

## Performance / Code Quality

- [ ] **Find should use selected mango index** (`find.go`) — `Find` currently
  ignores mango indexes entirely, scanning all documents and filtering in
  memory. `Explain` already selects the best index via `selectMangoIndex`;
  `Find` should do the same and use it to narrow the query.
- [ ] **Reduce caching** (`README.md`) — Reduce functions run on-demand with no
  intermediate result caching.
- [ ] **Mango SQL optimization** (`find.go`) — These selectors work via
  in-memory fallback but aren't translated to SQL: `$not`, `$nor`, `$nin`,
  `$regex`, `$mod`, `$all`, `$elemMatch`, `$type`, `$size`, `$allMatch`,
  `$keyMapMatch`.
- [ ] **Filter in Go instead of SQL** (`query.go:569`) — Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.
