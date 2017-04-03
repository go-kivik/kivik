# Key:

## Kivik components:

- ![Kivik API](images/api.png) : Supported by the Kivik Client API
- ![Kivik HTTP Server](images/http.png) : Supported by the Kivik HTTP Server
- ![Kivik Test Suite](images/tests.png) : Supported by the Kivik test suite
- ![CouchDB Logo](images/couchdb.png) : Supported by CouchDB backend
- ![PouchDB Logo](images/pouchdb.png) : Supported by PouchDB backend
- ![Memory Driver](images/memory.png) : Supported by Kivik Memory backend
- ![Filesystem Driver](images/filesystem.png) : Supported by the Kivik Filesystem backend

## API Functionality

- ✅ Yes : This feature is fully supported
- ☑️ Partial : This feature is partially supported
- ⍻ Emulated : The feature does not exist in the native driver, but is emulated.
- ？ Untested : This feature has been implemented, but is not yet fully tested.
- ⁿ/ₐ Not Applicable : This feature is supported, and doesn't make sense to emulate.
- ❌ No : This feature is supported by the backend, but there are no plans to add support to Kivik

<a name="authTable">

| Authentication Method | ![Kivik HTTP Server](images/http.png) | ![Kivik Test Suite](images/tests.png) | ![CouchDB](images/couchdb.png) | ![PouchDB](images/pouchdb.png) | ![Memory Driver](images/memory.png) | ![Filesystem Driver](images/filesystem.png) |
|--------------|:-------------------------------------:|:-------------------------------------:|:------------------------------:|:------------------------------:|:-----------------------------------:|:------------------------------------------:|
| HTTP Basic Auth    | ✅ | ✅ | ✅ | ✅<sup>[1](#pouchDbAuth)</sup> | ⁿ/ₐ | ⁿ/ₐ<sup>[2](#fsAuth)</sup>
| Cookie Auth        | ✅ | ✅ | ✅<sup>[3](#couchGopherJSAuth)</sup> |    | ⁿ/ₐ | ⁿ/ₐ<sup>[2](#fsAuth)</sup>
| Proxy Auth         |    |    |    |    | ⁿ/ₐ | ⁿ/ₐ<sup>[2](#fsAuth)</sup>
| OAuth 1.0          |    |    |    |    | ⁿ/ₐ | ⁿ/ₐ<sup>[2](#fsAuth)</sup>
| OAuth 2.0          |    |    | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ

### Notes

1. <a name="pouchDbAuth"> PouchDB Auth support is only for remote databases. Local databases rely on a same-origin policy.
2. <a name="fsAuth">The Filesystem driver depends on whatever standard filesystem permissions are implemented by your operating system. This means that you do have the option on a Unix filesystem, for instance, to set read/write permissions on a user/group level, and Kivik will naturally honor these, and report access denied errors as one would expect.
3. <a name="couchGopherJSAuth">Due to security limitations in the XMLHttpRequest spec, when compiling the standard CouchDB driver with GopherJS, CookieAuth will not work.

| API Endpoint | ![Kivik API](images/api.png) | ![Kivik HTTP Server](images/http.png) | ![Kivik Test Suite](images/tests.png) | ![CouchDB](images/couchdb.png) | ![PouchDB](images/pouchdb.png) | ![Memory Driver](images/memory.png) | ![Filesystem Driver](images/filesystem.png) |
|--------------|------------------------------|:-------------------------------------:|:-------------------------------------:|:------------------------------:|:------------------------------:|:-----------------------------------:|:------------------------------------------:|
| GET /        | ServerInfo()                 | ✅ | ✅ | ✅ | ✅ | ✅ | ✅ |
| GET /_active_tasks |                        |    |    |    | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| GET /_all_dbs      | AllDBs()               | ✅ | ✅ | ✅ | ☑️<sup>[1](#pouchAllDbs1),[2](#pouchAllDbs2),[3](pouchLocalOnly)</sup> | ✅ | ✅
| GET /_db_updates   | DBUpdates()            |    | ✅ | ✅ | ⁿ/ₐ |
| GET /_log          | Log()                  | ✅ | ✅ | ✅ | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| GET /_replicate
| GET /_restart      | ⁿ/ₐ                     |    |    | ❌<sup>[15](#notPublic)</sup> | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| GET /_stats        | ⁿ/ₐ                     |    |    | ❌<sup>[15](#notPublic)</sup> | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| GET /_utils        | ⁿ/ₐ                     |    |    |    | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| GET /_membership   | Membership()           | ❌<sup>[12](#kivikCluster)</sup> | ✅ | ✅<sup>[4](#couchMembership)</sup> | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ
| GET /favicon.ico   |                        | ✅ | ❌ | ❌ | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| POST /_session<sup>[6](#cookieAuth)</sup> | ⁿ/ₐ<sup>[13](#getSession)</sup> | ✅ | ✅ | ✅ | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| GET /_session<sup>[6](#cookieAuth)</sup> | ⁿ/ₐ<sup>[13](#getSession)</sup> | ☑️ | ✅ | ✅ | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| DELETE /_session<sup>[6](#cookieAuth)</sup> | ⁿ/ₐ<sup>[13](#getSession)</sup> | ✅ | ✅ | ✅ | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| * /_config         | Config()               |    | ✅ | ✅ | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| HEAD /{db}         | DBExists()             | ✅ | ✅ | ✅ | ✅<sup>[5](#pouchDBExists)</sup> | ✅ | ✅
| GET /{db}          | Info()                 |    | ✅ | ✅ | ✅
| PUT /{db}          | CreateDB()             | ✅ | ✅ | ✅ | ✅<sup>[5](#pouchDBExists)</sup> | ✅ | ✅
| DELETE /{db}       | DestroyDB()            |    | ✅ | ✅ | ✅<sup>[5](#pouchDBExists)</sup> | ✅ | ✅
| POST /{db}         | CreateDoc()            |    | ✅ | ✅ | ✅ |
| GET /{db}/_all_docs | AllDocs()             |    | ☑️<sup>[7](#todoConflicts),[9](#todoOrdering),[10](#todoLimit)</sup> | ✅ | ？ | ？ |
| POST /{db}/_all_docs | ⁿ/ₐ                   |    |    | ❌ | ❌ | ⁿ/ₐ | ⁿ/ₐ |
| POST /{db}/_bulk_docs | BulkDocs()          |    | ✅ | ✅ | ✅  |    |    |
| GET /{db}/_changes   | Changes()<sup>[8](#changesContinuous)</sup> |    | ✅ | ✅ | ✅ |    |    |
| POST /{db}/_changes  |                      |    |    |    | ⁿ/ₐ | ⁿ/ₐ | ⁿ/ₐ |
| POST /{db}/_compact  | Compact()            |    | ✅ | ✅ | ✅ |     |    |
| POST /{db}/_compact/{ddoc} | CompactView()  |    |    | ✅ | ⁿ/ₐ |    |    |
| POST /{db}/_ensure_full_commit | Flush()    | ✅ | ✅ | ✅ | ⁿ/ₐ | ⁿ/ₐ |    |
| POST /{db}/_view_cleanup | ViewCleanup()    |    | ✅ | ✅ | ✅ |     |    |
| GET /{db}/_security |                       |    | ✅ | ✅ | ⁿ/ₐ<sup>[14](#pouchPlugin)</sup>
| PUT /{db}/_security |                       |    | ✅ | ✅ | ⁿ/ₐ<sup>[14](#pouchPlugin)</sup>
| POST /{db}/_temp_view | ⁿ/ₐ                  | ⁿ/ₐ | ⁿ/ₐ| ⁿ/ₐ<sup>[16](#tempViews)</sup> | ⁿ/ₐ<sup>[17](#pouchTempViews)</sup> | ⁿ/ₐ | ⁿ/ₐ |
| POST /{db}/_purge   | ⁿ/ₐ                    |    |    | ❌<sup>[15](#notPublic)</sup> | ⁿ/ₐ |
| POST /{db}/_missing_revs | ⁿ/ₐ               |    |    | ❌<sup>[15](#notPublic)</sup> | ⁿ/ₐ |
| POST /{db}/_revs_diff | ⁿ/ₐ                  |    |    | ❌<sup>[15](#notPublic)</sup> | ⁿ/ₐ |
| GET /{db}/_revs_limit | RevsLimit()         |    | ✅ | ✅ | ☑️<sup>[3](#pouchLocalOnly)</sup> |
| PUT /{db}/_revs_limit | SetRevsLimit()      |    | ✅ | ✅ | ☑️<sup>[3](#pouchLocalOnly)</sup> |
| HEAD /{db}/{docid}  | Rev()                 |    | ✅ | ✅ | ⍻ |
| GET /{db}/{docid}   | Get()                 |    | ☑️<sup>[7](#todoConflicts),[11](#todoAttachments)</sup> | ✅ | ✅
| PUT /{db}/{docid}   | Put()                 |    | ☑️<sup>[11](#todoAttachments)</sup> | ✅ | ✅
| DELETE /{db}/{docid}| Delete()              |    | ✅ | ✅ | ✅
| COPY /{db}/{docid}  | Copy()                |    | ✅ | ✅ | ⍻
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

1. <a name="pouchAllDbs1"> PouchDB support for AllDbs depends on the
    [pouchdb-all-dbs plugin](https://github.com/nolanlawson/pouchdb-all-dbs).
2. <a name="pouchAllDbs2"> Unit tests broken in PouchDB due to an [apparent
    bug](https://github.com/nolanlawson/pouchdb-all-dbs/issues/25) in the
    pouchdb-all-dbs plugin.
3. <a name="pouchLocalOnly"> Supported for local PouchDB databases only. A work
    around may be possible in the future for remote databases.
4. <a name="couchMembership"> Available for CouchDB 2.0+ servers only.
5. <a name="pouchDBExists"> PouchDB offers no way to check for the existence of
 a local database without creating it, so `DBExists()` always returns true,
 `CreateDB()` does not return an error if the database already existed, and
 `DestroyDB()` does not return an error if the database does not exist.
6. <a name="cookieAuth"> See the CookieAuth section in the [Authentication methods table](#authTable)
7. <a name="todoConflicts"> **TODO:** Conflicts are not yet tested.
8. <a name="changesContinuous"> Changes feed operates in continuous mode only.
9. <a name="todoOrdering"> **TODO:** Ordering is not yet tested.
10. <a name="todoLimit"> **TODO:** Limits are not yet tested.
11. <a name="todoAttachments"> **TODO:** Attachments are not yet tested.
12. <a name="kivikCluster"> There are no plans at present to support clustering.
13. <a name="getSession"> Used for authentication, but not exposed directly to
    the client API.
14. <a name="pouchPlugin"> This feature is not available in the core PouchDB
    package. Support is provided in PouchDB plugins, so including optional
    support here may be possiblein the future.
15. <a name="notPublic"> This feature is not considered (by me, if nobody else)
    part of the public CouchDB API, so there are no (immediate) plans to
    implement support. If you feel this should change for a given feature,
    please create an issue and explain your reasons.
16. <a name="tempViews"> As of CouchDB 2.0, temp views are no longer supported,
    so I see no reason to support them in this library for older server versions.
    If you feel they should be supported, please create an issue and make your
    case.
17. <a name="pouchTempViews"> At present, PouchDB effectively supports temp
    views by calling [query](https://pouchdb.com/api.html#query_database) with
    a JS function. This feature is scheduled for removal from PouchDB (into a
    plugin), but until then, this functionality can still be used via the
    Query() method, by passing a JS function as an option.

## HTTP Status Codes

The CouchDB API prescribes some status codes which, to me, don't make a lot of
sense. This is particularly true of a few error status codes. It seems the folks
at [Cloudant](https://cloudant.com/) share my opinion, as they have chaned some
as well.

In particular, the CouchDB API returns a status 500 **Internal Server Error** for
quite a number of malformed requests.  Example: `/_uuids?count=-1` will return
500.  Cloudant and Kivik both return 400 **Bad Request** in this case, and in
many other cases as well, as this seems to better reflect the actual state.
