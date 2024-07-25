[![Go Reference](https://pkg.go.dev/badge/github.com/go-kivik/kivik/v4/pouchdb.svg)](https://pkg.go.dev/github.com/go-kivik/kivik/v4/pouchdb)

# Kivik PouchDB

PouchDB driver for [Kivik](https://github.com/go-kivik/pouchdb).

## Installation

Kivik 3.x and newer requires Go 1.11+, with Go modules enabled. At the time of
this writing, GopherJS still does not support Go modules (this is tracked at
[GopherJS Issue #855](https://github.com/gopherjs/gopherjs/issues/855)). Despite
this shortcoming of GopherJS, it is relatively straight forward to use the
standard Go toolchain as a dependency manager for GopherJS. I have written a
brief tutorial on this [here](https://jhall.io/posts/gopherjs-with-modules/),
with Kivik as an example.

## Usage

This package provides an implementation of the
[`github.com/go-kivik/kivik/v4/driver`](http://pkg.go.dev/github.com/go-kivik/kivik/v4/driver)
interface. You must import the driver and can then use the full
[`Kivik`](http://pkg.go.dev/github.com/go-kivik/kivik/v4) API.

```go
//go:build js

package main

import (
    "context"

    kivik "github.com/go-kivik/kivik/v4"
    _ "github.com/go-kivik/kivik/v4/pouchdb" // The PouchDB driver
)

func main() {
    client, err := kivik.New(context.TODO(), "pouch", "")
    // ...
}
```

This package is intended to run in a JavaScript runtime, such as a browser or
Node.js, and must be compiled with
[GopherJS](https://github.com/gopherjs/gopherjs). At runtime, the
[PouchDB](https://pouchdb.com/download.html) JavaScript library must also be
loaded and available.

## What license is Kivik released under?

This software is released under the terms of the Apache 2.0 license. See
LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
