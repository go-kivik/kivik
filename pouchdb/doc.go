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
Package pouchdb provides a [PouchDB] driver for [Kivik]. It must
be compiled with GopherJS, and requires that the PouchDB JavaScript library
is also loaded at runtime.

# General Usage

Use the `pouch` driver name when using this driver. The DSN should be the
blank string for local database connections, or a full URL, including any
required credentials, when connecting to a remote database.

	// +build js

	package main

	import (
		"context"

		kivik "github.com/go-kivik/kivik/v4"
		_ "github.com/go-kivik/kivik/v4/pouchdb" // PouchDB driver
	)

	func main() {
		client, err := kivik.New(context.TODO(), "pouch", "")
	// ...
	}

# Options

The PouchDB driver generally interprets [github.com/go-kivik/kivik/v4.Params]
keys and values as key/value pairs to pass to the relevant PouchDB method. In
general, PouchDB's key/value pairs are the same as the query parameters used
for CouchDB. Consult the [PouchDB API documentation] for the relevant methods for
any exceptions or special cases.

[PouchDB]: https://pouchdb.com/
[Kivik]: https://kivik.io/
[PouchDB API documentation]: https://pouchdb.com/api.html
*/
package pouchdb
