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
Package couchdb is a driver for connecting with a CouchDB server over HTTP.

# General Usage

Use the `couch` driver name when using this driver. The DSN should be a full
URL, likely with login credentials:

	import (
		kivik "github.com/go-kivik/kivik/v4"
		_ "github.com/go-kivik/kivik/v4/couchdb" // The CouchDB driver
	)

	client, err := kivik.New("couch", "http://username:password@127.0.0.1:5984/")

# Options

The CouchDB driver generally interprets [github.com/go-kivik/kivik/v4.Params]
keys and values as URL query parameters. Values of the following types will be
converted to their ppropriate string representation when URL-encoded:

  - bool
  - string
  - []string
  - int, uint, uint8, uint16, uint32, uint64, int8, int16, int32, int64

Passing any other type will return an error.

CouchDB also accepts a few special-case options defined in this package.

# Authentication

The CouchDB driver supports a number of authentication methods. For most uses,
you don't need to worry about authentication at all--just include authentication
credentials in your connection DSN:

	client, _ := kivik.New("couch", "http://user:password@localhost:5984/")

This will use Cookie authentication by default (or BasicAuth if compiled with
GopherJS).

To use one of the explicit authentication mechanisms, pass one of the
authentication options to [New]. For example:

	client, _ := kivik.New("couch", "http://localhost:5984/", couchdb.BasicAuth("bob", "secret"))

# Connection Options

Calls to [github.com/go-kivik/kivik/v4.New] may include options.
[OptionUserAgent] and [OptionHTTPClient] are the only options honored. Example:

	client, _ := kivik.New("couch", "http://localhost:5984/",
		couchdb.OptionUserAgent("My Custom User Agent String"),
		couchdb.OptionHTTPClient(&http.Client{
			Timeout: 15, // A shorter request timeout
		}),
	)

# Multipart PUT

Normally, to include an attachment in a CouchDB document, it must be base-64
encoded, which leads to increased network traffic and higher CPU load. CouchDB
also supports the option to [upload multiple attachments] in a single request
using the 'multipart/related' content type.

This is supported by the Kivik CouchDB driver as well. To take advantage of this
capability, the `doc` argument to [github.com/go-kivik/kivik/v4.DB.Put] must
be either:

  - a map of type `map[string]interface{}`, with a key called `_attachments',
    and value of type [github.com/go-kivik/kivik/v4.Attachments] or pointer
    to [github.com/go-kivik/kivik/v4.Attachments]
  - a struct, with a field having the tag `json:"_attachment"`, and the field
    having the type [github.com/go-kivik/kivik/v4.Attachments] or pointer to
    [github.com/go-kivik/kivik/v4.Attachments].

With this in place, the CouchDB driver will switch to `multipart/related` mode,
sending each attachment in binary format, rather than base-64 encoding it.

To function properly, each attachment must have an accurate
[github.com/go-kivik/kivik/v4.Attachment.Size] value. If
[github.com/go-kivik/kivik/v4.Attachment.Size] is unset, the entirely attachment
may be read to determine its size, prior to sending it over the network, leading
to delays and unnecessary I/O and CPU usage. The simplest way to ensure
efficiency is to use [NewAttachment]. See the documentation on that function
for proper usage.

Example:

	file, _ := os.Open("/path/to/photo.jpg")
	atts := &kivik.Attachments{
	    "photo.jpg": NewAttachment("photo.jpg", "image/jpeg", file),
	}
	doc := map[string]interface{}{
	    "_id":          "user123",
	    "_attachments": atts,
	}
	rev, err := db.Put(ctx, "user123", doc)

To disable the `multipart/related` capabilities entirely, you may pass the
[OptionNoMultipartPut] option. This will fallback to the default of
inline base-64 encoding the attachments.  Example:

	rev, err := db.Put(ctx, "user123", doc", couchdb.OptionNoMultipartPut())

# Server config for CouchDB 1.x

CouchDB allows querying the server configuration via the /_config endpoint. This
is supported with the [github.com/go-kivik/kivik/v4.DB.Config],
[github.com/go-kivik/kivik/v4.DB.ConfigSection],
[github.com/go-kivik/kivik/v4.DB.ConfigValue],
[github.com/go-kivik/kivik/v4.DB.SetConfigValue], and
[github.com/go-kivik/kivik/v4.DB.DeleteConfigKey] methods. Each of these methods
accepts a node argument, but CouchDB 1.x does not support per-node configuration
via this endpoint. If you are still using CouchDB 1.x, leave the node argument
empty, and this driver will revert to CouchDB-compatible operation.

[upload multiple attachments]: http://docs.couchdb.org/en/stable/api/document/common.html#creating-multiple-attachments
*/
package couchdb
