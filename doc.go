// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

/*
Package kivik provides a generic interface to CouchDB or CouchDB-like databases.

The kivik package must be used in conjunction with a database driver. The
officially supported drivers are:

  - CouchDB: https://github.com/go-kivik/kivik/v4/couchdb
  - PouchDB: https://github.com/go-kivik/kivik/v4/pouchdb (requires GopherJS)
  - MockDB: https://github.com/go-kivik/kivik/v4/mockdb

The Filesystem and Memory drivers are also available, but in early stages of
development, and so many features do not yet work:

  - FilesystemDB: https://github.com/go-kivik/kivik/v4/x/fsdb
  - MemoryDB: https://github.com/go-kivik/kivik/v4/x/memorydb

The kivik driver system is modeled after the standard library's `sql` and
`sql/driver` packages, although the client API is completely different due to
the different database models implemented by SQL and NoSQL databases such as
CouchDB.

The most methods, including those on [Client] and [DB] are safe to call
concurrently, unless otherwise noted.

# Working with JSON

CouchDB stores JSON, so Kivik translates Go data structures to and from JSON as
necessary. The conversion from Go data types to JSON, and vice versa, is
handled automatically according to the rules and behavior described in the
documentation for the standard library's [encoding/json] package.

# Options

Most client and database methods take optional arguments of the type [Option].
Multiple options may be passed, and latter options take precedence over earlier
ones, in case of a conflict.

[Params] and [Param] can be used to set options that are generally converted to
URL query parameters. Different backend drivers may also provide their own
unique options with driver-specific effects. Consult your driver's documentation
for specifics.

# Error Handling

Kivik returns errors that embed an HTTP status code. In most cases, this is the
HTTP status code returned by the server. The embedded HTTP status code may be
accessed easily using the HTTPStatus() method, or with a type assertion to
`interface { HTTPStatus() int }`. Example:

	if statusErr, ok := err.(interface{ HTTPStatus() int }); ok {
		status = statusErr.HTTPStatus()
	}

Any error that does not conform to this interface will be assumed to represent
a http.StatusInternalServerError status code.

# Authentication

For common usage, authentication should be as simple as including the authentication credentials in the connection DSN. For example:

	client, err := kivik.New("couch", "http://admin:abc123@localhost:5984/")

This will connect to `localhost` on port 5984, using the username `admin` and
the password `abc123`. When connecting to CouchDB (as in the above example),
this will use [cookie auth].

Depending on which driver you use, there may be other ways to authenticate, as
well. At the moment, the CouchDB driver is the only official driver which offers
additional authentication methods. Please refer to the [CouchDB package documentation]
for details.

[cookie auth]: https://docs.couchdb.org/en/stable/api/server/authn.html?highlight=cookie%20auth#cookie-authentication
[CouchDB package documentation]: https://pkg.go.dev/github.com/go-kivik/kivik/v4/couchdb
*/
package kivik // import "github.com/go-kivik/kivik/v4"
