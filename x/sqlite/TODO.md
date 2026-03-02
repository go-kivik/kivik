# SQLite Driver TODO

## Low Priority (polyfilled by kivik)

- [ ] **BulkDocs** (`db.go`) — kivik emulates via individual Put/CreateDoc.
- [ ] **Copy** (`db.go`) — kivik emulates via Get+Put.

## Performance / Code Quality

- [ ] **Reduce caching** (`README.md`) — Reduce functions run on-demand with no
  intermediate result caching.
- [ ] **Mango SQL optimization** (`find.go`) — These selectors work via
  in-memory fallback but aren't translated to SQL: `$not`, `$nor`, `$nin`,
  `$regex`, `$mod`, `$all`, `$elemMatch`, `$type`, `$size`, `$allMatch`,
  `$keyMapMatch`.
- [ ] **Filter in Go instead of SQL** (`query.go:569`) — Local and design
  document filtering during view updates is done in Go after fetching rows,
  rather than in the SQL query.
