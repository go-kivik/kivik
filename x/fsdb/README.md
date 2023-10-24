[![Go Reference](https://pkg.go.dev/badge/github.com/go-kivik/kivik/v4/x/fsdb.svg)](https://pkg.go.dev/github.com/go-kivik/kivik/v4/x/fsdb)

# Kivik FSDB

Experimental file system driver for [Kivik](https://github.com/go-kivik/kivik).

This driver stores documents on a plain filesystem.

# Status

This is very much a work in progress. Things are expected to change quickly.

## Usage

This package provides an implementation of the
[`github.com/go-kivik/kivik/driver`](http://pkg.go.dev/github.com/go-kivik/kivik/v4/driver)
interface. You must import the driver and can then use the full
[`Kivik`](http://pkg.go.dev/github.com/go-kivik/kivik/v4) API.

```go
package main

import (
    "context"

    "github.com/go-kivik/kivik/v4"
    _ "github.com/go-kivik/kivik/v4/x/fsdb" // The File system driver
)

func main() {
    client, err := kivik.New(context.TODO(), "fs", "")
    // ...
}
```

## License

This software is released under the terms of the Apache 2.0 license. See
LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
