[![Build Status](https://travis-ci.org/go-kivik/fsdb.svg?branch=master)](https://travis-ci.org/go-kivik/fsdb) [![Codecov](https://img.shields.io/codecov/c/github/go-kivik/fsdb.svg?style=flat)](https://codecov.io/gh/go-kivik/fsdb) [![GoDoc](https://godoc.org/github.com/go-kivik/fsdb?status.svg)](http://godoc.org/github.com/go-kivik/fsdb)

# Kivik FSDB

File system driver for [Kivik](https://github.com/go-kivik/fsdb) v4.

This driver stores documents on a plain filesystem.

# Status

This is very much a work in progress. Things are expected to change quickly.

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

    "github.com/go-kivik/kivik/v4"
    _ "github.com/go-kivik/fsdb/v4" // The File system driver
)

func main() {
    client, err := kivik.New(context.TODO(), "fs", "")
    // ...
}
```

## License

This software is released under the terms of the Apache 2.0 license. See
LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
