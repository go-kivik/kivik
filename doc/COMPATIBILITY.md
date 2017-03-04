# Key:

## Kivik components:

- ![Kivik API](images/api.png) : Supported by the Kivik Client API
- ![Kivik HTTP Server](images/http.png) : Supported by the Kivik HTTP Server
- ![Kivik Test Suite](images/tests.png) : Supported by the Kivik test suite
- ![CouchDB Logo](images/couchdb.png) : Supported by CouchDB backend
- ![PouchDB Logo](images/pouchdb.png) : Supported by PouchDB backend
- ![Memory Driver](images/memory.png) : Supported by Kivik Memory backend
- ![Filesystem Driver](images/filesystem.png) : Supported by the Kivik Filesystem backend

## Statuses

- ✅ Yes : This feature is fully supported
- ☑️ Partial : This feature is partially supported
- ？ Untested : This feature has been implemented, but is not yet fully tested.
- ⁿ/ₐ Not Applicable : This feature is supported, and doesn't make sense to emulate.
- ❌ No : This feature is supported by the backend, but there are no plans to add support to Kivik

<a name="authTable">

| Authentication Method | ![Kivik HTTP Server](images/http.png) | ![Kivik Test Suite](images/tests.png) | ![CouchDB](images/couchdb.png) | ![PouchDB](images/pouchdb.png) | ![Memory Driver](images/memory.png) | ![Filesystem Driver](images/filesystem.png) |
|--------------|:-------------------------------------:|:-------------------------------------:|:------------------------------:|:------------------------------:|:-----------------------------------:|:------------------------------------------:|
| HTTP Basic Auth    |    |    | ✅ | ✅<sup>[1](#pouchDbAuth)</sup> | ⁿ/ₐ | ⁿ/ₐ<sup>[2](#fsAuth)</sup>
| Cookie Auth        |    | ✅ | ✅<sup>[3](#couchGopherJSAuth)</sup> |    | ⁿ/ₐ | ⁿ/ₐ<sup>[2](#fsAuth)</sup>
| Proxy Auth         |    |    |    |    | ⁿ/ₐ | ⁿ/ₐ<sup>[2](#fsAuth)</sup>
| OAuth 1.0          |    |    |    |    | ⁿ/ₐ | ⁿ/ₐ<sup>[2](#fsAuth)</sup>

### Notes

1. <a name="pouchDbAuth"> PouchDB Auth support is only for remote databases. Local databases rely on a same-origin policy.
2. <a name="fsAuth">The Filesystem driver depends on whatever standard filesystem permissions are implemented by your operating system. This means that you do have the option on a Unix filesystem, for instance, to set read/write permissions on a user/group level, and Kivik will naturally honor these, and report access denied errors as one would expect.
3. <a name="couchGopherJSAuth">Due to security limitations in the XMLHttpRequest spec, when compiling the standard CouchDB driver with GopherJS, CookieAuth will not work.

| API Endpoint | ![Kivik API](images/api.png) | ![Kivik HTTP Server](images/http.png) | ![Kivik Test Suite](images/tests.png) | ![CouchDB](images/couchdb.png) | ![PouchDB](images/pouchdb.png) | ![Memory Driver](images/memory.png) | ![Filesystem Driver](images/filesystem.png) |
|--------------|------------------------------|:-------------------------------------:|:-------------------------------------:|:------------------------------:|:------------------------------:|:-----------------------------------:|:------------------------------------------:|
| GET /        | ServerInfo()                 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅
| GET /_active_tasks |                        |    |    |    | ⁿ/ₐ |
| GET /_all_dbs      | AllDBs()               | ✅ | ✅ | ✅ | ☑️<sup>[1](#pouchAllDbs1),[2](#pouchAllDbs2),[3](pouchAllDbs3)</sup> | ✅ | ✅
| GET /_db_updates
| GET /_log          | Log()                  |    |    | ✅ | ⁿ/ₐ
| GET /_replicate
| GET /_restart      |                        |    |    |    | ⁿ/ₐ
| GET /_stats
| GET /_utils        |                        |    |    |    | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ
| GET /_membership   | Membership()           |    |    | ✅<sup>[4](#couchMembership)</sup> | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ
| GET /favicon.ico   |                        |    | ❌ | ❌ | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ
| POST /_session<sup>[6](#cookieAuth)</sup> | |    | ✅ | ✅ | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ
| GET /_session<sup>[6](#cookieAuth)</sup> |  |    |    |    | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ
| DELETE /_session<sup>[6](#cookieAuth)</sup> | |    |    |    | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ
| GET /_config
| GET /_config/{section}
| GET /_config/{section}/{key}
| PUT /_config/{section}/{key}
| DELETE /_config/{section}/{key}
| HEAD /{db}         | DBExists()             | ✅ |    | ✅ | ✅<sup>[5](#pouchDBExists)</sup> | ✅ | ✅
| GET /{db}
| PUT /{db}          | CreateDB()             | ✅ |    | ✅ | ✅<sup>[5](#pouchDBExists)</sup> | ✅ | ✅
| DELETE /{db}       | DestroyDB()            |    |    | ✅ | ✅<sup>[5](#pouchDBExists)</sup> | ✅ | ✅
| POST /{db}
| GET /{db}/_all_docs | AllDocs()             |    | ☑️<sup>[7](#todoConflicts),[8](#todoIncludeDocs),[9](#todoOrdering),[10](#todoLimit)</sup> | ✅ | ？ | ？ |
| POST /{db}/_all_docs
| POST /{db}/_bulk_docs
| GET /{db}/_changes
| POST /{db}/_changes
| POST /{db}/_compact
| POST /{db}/_compact/{ddoc}
| POST /{db}/_ensure_full_commit
| POST /{db}/_view_cleanup
| GET /{db}/_security
| PUT /{db}/_security
| POST /{db}/_temp_view
| POST /{db}/_purge
| POST /{db}/_missing_revs
| POST /{db}/_revs_diff
| GET /{db}/_revs_limit
| PUT /{db}/_revs_limit
| HEAD /{db}/{docid}
| GET /{db}/{docid}   | Get()                 |    | ☑️<sup>[7](#todoConflicts),[11](#todoAttachments)</sup> | ✅ |
| PUT /{db}/{docid}   | Put()                 |    | ☑️<sup>[11](#todoAttachments)</sup> | ✅ |
| DELETE /{db}/{docid}
| COPY /{db}/{docid}
| HEAD /{db}/{docid}/{attname}
| GET /{db}/{docid}/{attname}
| PUT /{db}/{docid}/{attname}
| DELETE /{db}/{docid}/{attname}
| HEAD /{db}/_design/{ddoc}
| GET /{db}/_design/{ddoc}
| PUT /{db}/_design/{ddoc}
| DELETE /{db}/_design/{ddoc}
| COPY /{db}/_design/{ddoc}
| HEAD /{db}/_design/{ddoc}/{attname}
| GET /{db}/_design/{ddoc}/{attname}
| PUT /{db}/_design/{ddoc}/{attname}
| DELETE /{db}/_design/{ddoc}/{attname}
| GET /{db}/_design/{ddoc}/_info
| GET /{db}/_design/{ddoc}/_view/{view}
| POST /{db}/_design/{ddoc}/_view/{view}
| GET /{db}/_design/{ddoc}/_show/{func}
| POST /{db}/_design/{ddoc}/_show/{func}
| GET /{db}/_design/{ddoc}/_show/{func}/{docid}
| POST /{db}/_design/{ddoc}/_show/{func}/{docid}
| GET /{db}/_design/{ddoc}/_list/{func}/{view}
| POST /{db}/_design/{ddoc}/_list/{func}/{view}
| GET /{db}/_design/{ddoc}/_list/{func}/{other-ddoc}/{view}
| POST /{db}/_design/{ddoc}/_list/{func}/{other-ddoc}/{view}
| POST /{db}/_design/{ddoc}/_update/{func}
| PUT /{db}/_design/{ddoc}/_update/{func}/{docid}
| ANY /{db}/_design/{ddoc}/_rewrite/{path}
| GET /{db}/_local/{docid}
| PUT /{db}/_local/{docid}
| DELETE /{db}/_local/{docid}
| COPY /{db}/_local/{docid}

### Notes

1. <a name="pouchAllDbs1"> PouchDB support for AllDbs depends on the [pouchdb-all-dbs plugin](https://github.com/nolanlawson/pouchdb-all-dbs).
2. <a name="pouchAllDbs2"> Unit tests broken in PouchDB due to an [apparent bug](https://github.com/nolanlawson/pouchdb-all-dbs/issues/25) in the pouchdb-all-dbs plugin.
3. <a name="pouchAllDbs3"> Does not work for remote PouchDB connections, due to a limitation in the `pouchdb-all-dbs` plugin. Perhaps a workaround will be possible in the future.
4. <a name="couchMembership"> Available for CouchDB 2.0+ servers only.
5. <a name="pouchDBExists"> PouchDB offers no way to check for the existence of a local database
 without creating it, so `DBExists()` always returns true, `CreateDB()` does not return an error
 if the database already existed, and `DestroyDB()` does not return an error if the database does
 not exist.
6. <a name="cookieAuth"> See the CookieAuth section in the [Authentication methods table](#authTable)
7. <a name="todoConflicts"> **TODO:** Conflicts are not yet tested.
8. <a name="todoIncludeDocs"> **TODO:** include_docs is not yet tested.
9. <a name="todoOrdering"> **TODO:** Ordering is not yet tested.
10. <a name="todoLimit"> **TODO:** Limits are not yet tested.
11. <a name="todoAttachments"> **TODO:** Attachments are not yet tested.
