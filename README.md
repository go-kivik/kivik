[![Build Status](https://travis-ci.org/flimzy/kivik.svg?branch=master)](https://travis-ci.org/flimzy/kivik) [![GoDoc](https://godoc.org/github.com/flimzy/kivik?status.png)](http://godoc.org/github.com/flimzy/kivik)

# Kivik

Package kivik provides a generic interface to CouchDB or CouchDB-like databases.

The kivik package must be used in conjunction with a database driver.

The kivik driver system is modeld after the standard library's [sql](https://golang.org/pkg/database/sql/)
and [sql/driver](https://golang.org/pkg/database/sql/driver/) packages, although
the client API is completely different due to the different database models
implemented by SQL and NoSQL databases such as CouchDB.

# Frequently Asked Questions

Nobody has ever asked me any of these questions, so they're probably better called
"Never Asked Questions" or possibly "Imagined Questions."

## Why another CouchDB client API?

A few reasons:

1. I'm not happy with any of the existing ones. The best existing one seems to
be https://github.com/fjl/go-couchdb, but it has several shortcomings:

  - It's not being actively developed. The author has not responded to any pull
    requests or issues for over two years, from what I can tell.
  - It [doesn't have an open source license](https://github.com/fjl/go-couchdb/issues/15).
    In fact, it has no license at all, which, strictly speaking, means it should
    not ever be used by anyone without permission (which is presently impossible
    to acquire, thanks to the author not responding to issues).
  - It [doesn't support CouchDB 2.0](https://github.com/fjl/go-couchdb/issues/14).
    The API differences between 1.6.1 and 2.0 are minor, but they do exist.

2. I want a single client API that works with both CouchDB and [PouchDB](https://pouchdb.com/).
I have previously written [go-pouchdb](https://github.com/flimzy/go-pouchdb), which is
a GopherJS wrapper around the PouchDB library. And I modeled my go-pouchdb API
after fjl/go-couchdb. But they're still separate libraries. One of kivik's main
design goals is to allow the exact same API to access both, in the same way
that the standard library's [sql](https://golang.org/pkg/database/sql/) package
allows the same API to access MySQL, PostgreSQL, SQLite, or any other SQL database.

3. I want an unambiguous, open source license. This software is released under
the Apache 2.0 license. See the included LICENSE.md file for details.

4. I want the ability to mock CouchDB connections for testing. This is possible
with the sql / sql/driver approach, by implementing a mock driver, but was not
possible with any existing CouchDB client libraries. This library intends to
remedy that.

5. In the future term, I intend to expand kivik to support a 'serve' mode. This
will allow running a minimal stand-alone CouchDB server or proxy, for the purpose
of testing. My personal goal is to run a kivik server with a memory backend, so
that I can test a CouchDB app, or PouchDB syncing, against my kivik server,
without the administrative overhead of installing a full-fledged CouchDB instance.
If successful, this opens up additional possibilities of creating simple CouchDB
proxy servers, which could be quite interesting.

## What is the development status?

Kivik is in early stages of development. At the moment, it barely does anything
at all. My development goals are, in rough order of priority:

1. Complete the 'kivik' client API.
2. Complete the CouchDB driver.
3. Complete the PouchDB driver.
4. Complete the Memory driver.
5. Write a 'serve' mode.

1-4 are being done roughly in parallel, although the features of the various
drivers are not entirely the same.

## Why the name "Kivik"?

[Kivik](http://www.ikea.com/us/en/catalog/categories/series/18329/) is a line
of sofas (couches) from IKEA. And in the spirit of IKEA, and build-your-own
furniture, Kivik aims to allow you to "build your own" CouchDB client, server,
and proxy applications.

## What license is Kivik released under?

This software is released under the terms of the Apache 2.0 license. See
LICENCE.md, or read the [full license](http://www.apache.org/licenses/LICENSE-2.0).
