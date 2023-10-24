[![Go Reference](https://pkg.go.dev/badge/github.com/go-kivik/kivik/v4/couchdb.svg)](https://pkg.go.dev/github.com/go-kivik/kivik/v4/couchdb)

# Kivik CouchDB

CouchDB driver for [Kivik](https://github.com/go-kivik/kivik).

## Usage

This package provides an implementation of the
[`github.com/go-kivik/kivik/v4/driver`](http://pkg.go.dev/github.com/go-kivik/kivik/v4/driver)
interface. You must import the driver and can then use the full
[`Kivik`](http://pkg.go.dev/github.com/go-kivik/kivik/v4) API.

```go
package main

import (
    "context"

    kivik "github.com/go-kivik/kivik/v4"
    _ "github.com/go-kivik/kivik/v4/couchdb" // The CouchDB driver
)

func main() {
    client, err := kivik.New(context.TODO(), "couch", "")
    // ...
}
```

## License

This software is released under the terms of the Apache 2.0 license. See
LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
