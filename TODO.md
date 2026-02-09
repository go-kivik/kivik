# Kivik Project TODO

## Confirmed Bugs

- [ ] **`Test_ping_RunE/network_error` failing in `cmd/kivik/cmd`**
  Test expects a `connection refused` error and exit code 69, but gets
  `Server down` and exit code 14. Likely something listening on port 9999
  or an environment-dependent behavior change. Snapshot needs regeneration
  or test logic needs updating.

## Feature Gaps / Incomplete Implementations

- [ ] **ProxyDB has unimplemented methods that panic**
  `PutAttachment`, `GetAttachment`, `CreateDoc`, `Delete`,
  `DeleteAttachment`, `Put` all panic with "should never be called".

- [ ] **MemoryDB missing attachment support**
  `x/memorydb/db.go:84` — `TODO: Add support for storing attachments`

- [ ] **SQLite driver missing many optional interfaces**
  No `BulkDocs`, `Copy`, `DesignDocs`, `LocalDocs`, `Purge`, `Security`,
  `Flush`, `Config`, `Session`, `Replication`, `Cluster` support.
  Some return stub "not implemented" errors, others simply don't implement
  the interface.

- [ ] **Searcher interface inconsistency**
  `driver/search.go` uses `map[string]interface{}` for options instead of
  `driver.Options` like every other interface method. Neither CouchDB nor
  SQLite implements `Searcher`.

- [ ] **No `ClientCloser` implementation**
  Neither CouchDB nor SQLite clients implement `Close()`, meaning users
  can't properly clean up client resources.

## Test Coverage Gaps

### Packages with zero test files

- [ ] `cmd/kivik/log` — CLI logging
- [ ] `int/mock` — 12 source files of mocking infrastructure
- [ ] `x/proxydb` — proxy database
- [ ] `x/server/auth` — server authentication
- [ ] `x/fsdb/cdb/decode` — document decoding

### Under-tested areas

- [ ] `cmd/kivik/output/` — all 4 output formatters (yaml, json, raw, gotmpl) have no tests
- [ ] `x/server/` — 12 untested source files including auth, security, middleware
- [ ] `driver/` — 10 interface files with no direct tests

## Go Modernization Opportunities

Root module is Go 1.20 (GopherJS constraint). Sub-modules `x/sqlite` (1.22)
and `x/pg` (1.24) can use newer features.
