[![Go Reference](https://pkg.go.dev/badge/github.com/go-kivik/kivik/x/sqlite/v4.svg)](https://pkg.go.dev/github.com/go-kivik/kivik/x/sqlite/v4)

# Kivik SQLite backend

SQLite-backed driver for [Kivik](https://github.com/go-kivik/kivik).

## Usage

This package provides a (partial, experimental) implementation of the
[`github.com/go-kivik/kivik/v4/driver`](http://pkg.go.dev/github.com/go-kivik/kivik/v4/driver)
interface. You must import the driver and can then use the
[`Kivik`](http://pkg.go.dev/github.com/go-kivik/kivik/v4) API.

```go
package main

import (
    "context"

    kivik "github.com/go-kivik/kivik/v4"
    _ "github.com/go-kivik/kivik/x/sqlite/v4" // The SQLite driver
)

func main() {
    client, err := kivik.New(context.TODO(), "sqlite", "/path/to/file.db")
    // ...
}
```

## Why?

The primary intended purpose of this driver is for testing. The goal is to allow
you to test your CouchDB apps without relying on a full-fledged CouchDB server.

## Status

This driver is incomplete, experimental, and under rapid development.

## Incompatibilities

The SQLite implementation of CouchDB is incompatible with the CouchDB specification in a few subtle ways, which are outlined here:

- The Collation order supported by Go is slightly different than that described by the [CouchDB documentation](https://docs.couchdb.org/en/stable/ddocs/views/collation.html#collation-specification). [See the GoDoc for details](https://pkg.go.dev/github.com/go-kivik/kivik/v4/x/collate#pkg-overview).
- While `map` functions are treated roughly the same as in CouchDB (that is, they are called when the view is first requested, then incremental updates made after that) `reduce` functions are always run on demand at the moment, with no intermediate caching. For small databases as in test scenarios, the primary use case for this library, this should be fine. But in the long run, this should be improved to make querying reduce views more efficient.

## License

This software is released under the terms of the Apache 2.0 license. See
LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
