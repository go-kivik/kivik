[![Build Status](https://gitlab.com/go-kivik/kivik/badges/master/pipeline.svg)](https://gitlab.com/go-kivik/kivik/pipelines) [![Codecov](https://img.shields.io/codecov/c/github/go-kivik/kivik.svg?style=flat)](https://codecov.io/gh/go-kivik/kivik) [![Go Report Card](https://goreportcard.com/badge/github.com/go-kivik/kivik)](https://goreportcard.com/report/github.com/go-kivik/kivik) [![Go Reference](https://pkg.go.dev/badge/github.com/go-kivik/kivik/v4.svg)](https://pkg.go.dev/github.com/go-kivik/kivik/v4) [![Website](https://img.shields.io/website-up-down-green-red/http/kivik.io.svg?label=website&colorB=007fff)](http://kivik.io)

# Kivik

Package kivik provides a common interface to CouchDB or CouchDB-like databases.

The kivik package must be used in conjunction with a database driver.

The kivik driver system is modeled after the standard library's [database/sql](https://golang.org/pkg/database/sql/)
and [sql/driver](https://golang.org/pkg/database/sql/driver/) packages, although
the client API is completely different due to the different database models
implemented by SQL and NoSQL databases such as CouchDB.

# Versions

You are browsing the current **stable** branch of Kivik, v4. If you are upgrading
from the previous stable version, v3, [read the list of breaking changes](#changes-from-3x-to-4x).

Example configuration for common dependency managers follow.

## Go Modules

Kivik 3.x and later supports Go modules, which is the recommended way to use it
for Go version 1.11 or newer. Kivik 4.x only supports Go 1.17 and later. If your
project is already using Go modules, simply fetch the desired version:

```shell
go get github.com/go-kivik/kivik/v4
```

# Installation

Install Kivik as you normally would for any Go package:

    go get -u github.com/go-kivik/kivik/v4

This will install the main Kivik package and the CouchDB database driver. Three officially supported drivers are shipped with this Go module:

- CouchDB: [github.com/go-kivik/kivik/v4/couchdb](https://pkg.go.dev/github.com/go-kivik/kivik/v4/couchdb)
- PouchDB: [github.com/go-kivik/kivik/v4/pouchdb](https://pkg.go.dev/github.com/go-kivik/kivik/v4/pouchdb) (requires GopherJS)
- MockDB: [github.com/go-kivik/kivik/v4/mockdb](https://pkg.go.dev/github.com/go-kivik/kivik/v4/mockdb)

In addition, there are partial/experimental drivers available:

- FilesystemDB: [github.com/go-kivik/kivik/v4/x/fsdb](https://pkg.go.dev/github.com/go-kivik/kivik/v4/x/fsdb)
- MemoryDB: [github.com/go-kivik/kivik/v4/x/memorydb](https://pkg.go.dev/github.com/go-kivik/kivik/v4/x/memorydb)
- SQLite: [github.com/go-kivik/kivik/x/sqlite/v4](https://pkg.go.dev/github.com/go-kivik/kivik/x/sqlite/v4)

# CLI

Consult the [CLI README](https://github.com/go-kivik/kivik/blob/master/cmd/kivik/README.md) for full details on the `kivik` CLI tool.

# Example Usage

Please consult the the [package documentation](https://pkg.go.dev/github.com/go-kivik/kivik/v4)
for all available API methods, and a complete usage documentation, and usage examples.

```go
package main

import (
    "context"
    "fmt"

    kivik "github.com/go-kivik/kivik/v4"
    _ "github.com/go-kivik/kivik/v4/couchdb" // The CouchDB driver
)

func main() {
    client, err := kivik.New("couch", "http://localhost:5984/")
    if err != nil {
        panic(err)
    }

    db := client.DB("animals")

    doc := map[string]interface{}{
        "_id":      "cow",
        "feet":     4,
        "greeting": "moo",
    }

    rev, err := db.Put(context.TODO(), "cow", doc)
    if err != nil {
        panic(err)
    }
    fmt.Printf("Cow inserted with revision %s\n", rev)
}
```

# Frequently Asked Questions

Nobody has ever asked me any of these questions, so they're probably better called
"Never Asked Questions" or possibly "Imagined Questions."

## Why another CouchDB client API?

I had a number of specific design goals when creating this package:

- Provide a generic database API for a variety of CouchDB-like databases. The previously existing drivers for CouchDB had patchy support for different versions of CouchDB, and different subsets of functionality.
- Work equally well with CouchDB 1.6, 2.x, 3.x, and any future versions, as well as PouchDB.
- Be as Go-idiomatic as possible.
- Be unambiguously open-source. Kivik is released under the Apache license, same as CouchDB and PouchDB.
- Separate out the basic implementation of a database driver (implementing the `kivik/driver` interfaces) vs the implementation of all the user-level types and convenience methods. It ought to be reasonably easy to create a new driver, for testing, mocking, implementing a new backend data storage system, or talking to other CouchDB-like databases.

## What are Kivik's requirements?

Kivik's test suite is automatically run on Linux for every pull request, but
should work on all supported Go architectures. If you find it not working for
your OS/architecture, please submit a bug report.

Below are the compatibility targets for specific runtime and database versions.
If you discover a bug affecting any of these supported environments, please let
me know by submitting a bug report via GitHub.

- **Go** Kivik 4.x aims for full compatibility with all stable releases of Go
  from 1.13. For Go 1.7 or 1.8 you can use [Kivik 1.x](https://github.com/go-kivik/kivik/tree/v1).
  For Go 1.9 through 1.12, you can use [Kivik 3.x](https://github.com/go-kivik/kivik/tree/v3).
- **CouchDB** The Kivik 4.x CouchDB driver aims for compatibility with all
  stable releases of CouchDB from 2.x.
- **GopherJS** GopherJS always requires the latest stable version of Go, so
  building Kivik with GopherJS has this same requirement.
- **PouchDB** The Kivik 4.x PouchDB driver aims for compatibility with all
  stable releases of PouchDB from 8.0.0.

## What is the development status?

Kivik 4.x is stable, and suitable for production use.

## Why the name "Kivik"?

[Kivik](http://www.ikea.com/us/en/catalog/categories/series/18329/) is a line
of sofas (couches) from IKEA. And in the spirit of IKEA, and build-your-own
furniture, Kivik aims to allow you to "build your own" CouchDB client, server,
and proxy applications.

## What license is Kivik released under?

Kivik is Copyright 2017-2023 by the Kivik contributors, and is released under the
terms of the Apache 2.0 license. See [LICENCE](LICENSE) for the full text of the
license.

### Changes from 3.x to 4.x

This is a partial list of breaking changes between 3.x and 4.x

- Options are no longer a simple `map[string]interface{}`, but are rather functional parameters. In most cases, you can just use `kivik.Param(key, value)`, or `kivik.Params(map[string]interface{}{key: value})` as a replacement. Some shortcuts for common params now exist, and driver-specific options may work differently. Consult the GoDoc.
- The `Authenticate` method has been removed. Authentication is now handled via option parameters.
- The CouchDB, PouchDB, and MockDB drivers, and the experimental FilesystemDB and MemoryDB drivers, have been merged with this repo, rather than being hosted in separate repos. For v3 you would have imported `github.com/go-kivik/couchdb/v3`, for example. With v4, you instead use `github.com/go-kivik/kivik/v4/couchdb` for CouchDB, or `github.com/go-kivik/kivik/v4/x/fsdb` for the experimental FilesystemDB driver.
- The return type for queries has been significantly changed.
  - In 3.x, queries returned a `*Rows` struct. Now they return a `*ResultSet`.
  - The `Offset()`, `TotalRows()`, `UpdateSeq()`, `Warning()` and `Bookmark()` methods have been removed, and replaced with the `ResultMetadata` type which is accessed via the `Metadata()` method. See [issue #552](https://github.com/go-kivik/kivik/issues/552).
  - Calling most methods on `ResultSet` will now work after closing the iterator.
  - The new `ResultSet` type supports multi-query mode, which is triggered by calling `NextResultSet` before `Next`.
  - `Key`, `ID`, `Rev`, `Attachments` all now return row-specific errors, and `ScanKey` may successfully decode while also returning a row-specific error.
- The `Changes` type has been changed to semantically match the `ResultSet` type. Specifically, the `LastSeq()` and `Pending()` methods have been replaced by the `Metadata()` method.
- The `DBUpdates()` and `Changes()` methods now defer errors to the iterator, for easier chaining and consistency with other iterators.
- `DB.BulkDocs()` no longer returns an iterator, but rather an array of all results.
- `Get` now returns a simpler `*Result` type than before.
- `GetMeta` has been replaced with `GetRev`, and no longer claims to return the document size. The document size was never _really_ the document size, rather it is the `Content-Length` field of the HTTP response, which can vary depending on query parameters, making its use for determining document size dubious at best.
- The `StatusCode() int` method on errors has been renamed to `HTTPStatus() int`, to be more descriptive. The related package function `StatusCode(error) int` has also been renamed to `HTTPStatus(error) int` to match.
- `Client.Close()` and `DB.Close()` now block until any relevant calls have returned.
- `Client.Close()` and `DB.Close()` no longer take a `context.Context` value. These operations cannot actually be canceled anyway, by the one driver that uses them (PouchDB); it only stops waiting. It makes more senes to make these functions blocking indefinitely, especially now that they wait for client requests to finish, and let the caller stop waiting if it wants to.

#### CouchDB specific changes

- The `SetTransport` authentication method has been removed, as a duplicate of [couchdb.OptionHTTPClient](https://pkg.go.dev/github.com/go-kivik/kivik/v4/couchdb#OptionHTTPClient).
- Options passed to Kivik functions are now functional options, rather than a map of string to empty interface. As such, many of the options have changed. Consult the relevant GoDoc.

#### New features and additions

- Kivik now ships with the `kivik` command line tool (previously part of the `github.com/go-kivik/xkivik` repository).
- The new [Replicate](https://pkg.go.dev/github.com/go-kivik/kivik/v4#Replicate) function allows replication between arbitrary databases, such as between CouchDB and a directory structure using the FilesystemDB.

## What projects currently use Kivik?

If your project uses Kivik, and you'd like to be added to this list, create an
issue or submit a pull request.

- [Cayley](https://github.com/cayleygraph/cayley) is an open-source graph
  database. It uses Kivik for the CouchDB and PouchDB storage backends.
