[![Go Reference](https://pkg.go.dev/badge/github.com/go-kivik/kivik/v4/x/memorydb.svg)](https://pkg.go.dev/github.com/go-kivik/kivik/v4/x/memorydb)

# Kivik MemoryDB

Experimental memory driver for [Kivik](https://github.com/go-kivik/kivik).

This driver stores documents in memory only, and is intended for testing purposes only. Not all Kivik features are or will be supported. This package is still under active development.

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
    _ "github.com/go-kivik/kivik/v4/x/memorydb" // The Memory driver
)

func main() {
    client, err := kivik.New(context.TODO(), "memory", "")
    // ...
}
```

## License

This software is released under the terms of the Apache 2.0 license. See
LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
