[![Build Status](https://travis-ci.org/go-kivik/memorydb.svg?branch=master)](https://travis-ci.org/go-kivik/memorydb) [![Codecov](https://img.shields.io/codecov/c/github/go-kivik/memorydb.svg?style=flat)](https://codecov.io/gh/go-kivik/memorydb) [![GoDoc](https://godoc.org/github.com/go-kivik/memorydb?status.svg)](http://godoc.org/github.com/go-kivik/memorydb)

# Kivik MemoryDB

Memory driver for [Kivik](https://github.com/go-kivik/memorydb).

This driver stores documents in memory only, and is intended for testing purposes only. Not all Kivik features are or will be supported. This package is still under active development.

## Usage

This package provides an implementation of the
[`github.com/go-kivik/kivik/driver`](http://godoc.org/github.com/go-kivik/kivik/driver)
interface. You must import the driver and can then use the full
[`Kivik`](http://godoc.org/github.com/go-kivik/kivik) API. Please consult the
[Kivik wiki](https://github.com/go-kivik/kivik/wiki) for complete documentation
and coding examples.

```go
package main

import (
    "context"

    "github.com/go-kivik/kivik"
    _ "github.com/go-kivik/memorydb" // The Memory driver
)

func main() {
    client, err := kivik.New(context.TODO(), "memory", "")
    // ...
}
```

## License

This software is released under the terms of the Apache 2.0 license. See
LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
