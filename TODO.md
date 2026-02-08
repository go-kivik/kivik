# Kivik Project TODO

## Confirmed Bugs

- [ ] **Response body leak in `couchdb/scheduler.go:99-119`**
  `schedulerSupported()` calls `c.DoReq()` but never closes `resp.Body`. Every
  call to check scheduler support leaks the response body.

- [ ] **Response body leak in `couchdb/client.go:89-90`**
  In `DBUpdates()`, if `chttp.ResponseError(resp)` returns an error, the
  function returns without closing `resp.Body`.

- [ ] **Typo in `cmd/kivik/output/output.go:72`**
  `panic("no formatters regiestered")` — "regiestered" should be "registered".

- [ ] **Panic instead of error in `couchdb/client.go:142`**
  `newUpdates()` has `panic("unknown")` as the default case for unrecognized
  feed types. Should return an error.

- [ ] **Panic instead of error in `couchdb/db.go:59`**
  `d.path()` panics with `"THIS IS A BUG: d.path failed"` on URL parse errors.
  Should return an error.

## Potential Concurrency Issues

- [ ] **Data race risk in `couchdb/db.go:640-672`**
  `newMultipartAttachments()` shares an `err` variable between a goroutine and
  the main function. While `sync.WaitGroup` provides a happens-before guarantee,
  the pattern is fragile and unconventional.

- [ ] **Goroutine leak in `couchdb/chttp/chttp.go:220-235`**
  `compressBody()` spawns a goroutine writing to an `io.Pipe`. If the pipe
  reader is abandoned before the goroutine finishes, the goroutine hangs
  indefinitely.

- [ ] **Goroutine leak in `couchdb/db.go:813-830`**
  `replaceAttachments()` has the same pipe-writer goroutine leak pattern.

## Unsafe Type Assertions (Panic Risk)

Internal interfaces, so low risk, but bare type assertions without the `ok`
check will panic at runtime if the wrong type is passed:

- [ ] `couchdb/rows.go:45,56,95,115` — parser methods
- [ ] `couchdb/changes.go:67,72` — changes parser
- [ ] `couchdb/client.go:124` — updates parser
- [ ] `couchdb/db.go:1112` — allDocs iterator

## Production Panics Worth Addressing

Beyond the confirmed bugs above, these panics in non-test, non-experimental
code should be converted to error returns:

- [ ] `options.go:92` — `"kivik: unknown option type: %T"`
- [ ] `couchdb/chttp/trace.go:64` — `"nil trace"`
- [ ] `pouchdb/replicationEvents.go:46,58` — panics on time parse errors
- [ ] `pouchdb/replicationEvents.go:110` — panics on unexpected replication event
- [ ] `cmd/kivik/cmd/get_db.go:59` — `panic(err)` on JSON marshal failure

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

### `strings.CutPrefix` / `strings.CutSuffix` (Go 1.20)

Replace `HasPrefix`+`TrimPrefix` and `HasSuffix`+`TrimSuffix` pairs:

- [ ] `couchdb/chttp/encode.go:33-34` — CutPrefix on docID prefix
- [ ] `kiviktest/testcontainers_js.go:138-140` — CutPrefix on "Listening on "
- [ ] `x/kivikd/authdb/confadmin/confadmin.go:68-72` — CutPrefix on hash prefix
- [ ] `cmd/kivik/config/config.go:106-108` — CutSuffix on "/_design"
- [ ] `x/fsdb/changes.go:68-69` — CutSuffix on file extension

### `strings.Cut` / `bytes.Cut` (Go 1.18)

Replace `SplitN(s, sep, 2)` with `Cut`:

- [ ] `cmd/kivik/cmd/put_config.go:66` — SplitN(key, "/", 2)
- [ ] `cmd/kivik/cmd/get_config.go:73` — SplitN(key, "/", 2)
- [ ] `cmd/kivik/cmd/delete_config.go:61` — SplitN(key, "/", 2)
- [ ] `cmd/kivik/output/output.go:111` — SplitN(format, "=", 2)
- [ ] `cmd/kivik/config/config.go:507` — SplitN(dsn, "//", 2)
- [ ] `x/options/options.go:246` — SplitN(filter, "/", 2)
- [ ] `x/memorydb/db_test.go:274` — SplitN(rev, "-", 2)
- [ ] `couchdb/replication.go:179` — SplitN(replicationID, "+", 2)
- [ ] `x/fsdb/cdb/revid.go:69` — SplitN(rev, "-", 2)
- [ ] `x/fsdb/cdb/revid.go:38-48` — bytes.SplitN with Contains guard → bytes.Cut
- [ ] `x/sqlite/options.go:227` — SplitN(filter, "/", 2) *(separate module)*
- [ ] `x/sqlite/json.go:54` — SplitN(rev, "-", 2) *(separate module)*

### `os.IsNotExist` → `errors.Is(err, fs.ErrNotExist)` (Go 1.16)

`errors.Is` correctly unwraps wrapped errors; `os.IsNotExist` does not.

- [ ] `x/fsdb/fs.go:148`
- [ ] `x/fsdb/put.go:41`
- [ ] `x/fsdb/errors.go:29`
- [ ] `x/fsdb/cdb/errors.go:30, 44`
- [ ] `x/fsdb/cdb/fs.go:123`
- [ ] `x/fsdb/cdb/security.go:44, 60`
- [ ] `x/fsdb/cdb/revision.go:161, 173`
- [ ] `x/fsdb/cdb/decode/decode.go:51`
- [ ] `x/kivikd/couchserver/favicon.go:43`
- [ ] `cmd/kivik/config/config.go:194`

### `interface{}` → `any` (Go 1.18)

~2,397 occurrences across ~266 files. Mechanical replacement via
`gofmt -r 'interface{} -> any' -w .`. Verify generated code templates in
`mockdb/gen/` are also updated so regeneration doesn't revert the change.
