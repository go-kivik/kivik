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

package couchdb

import (
	"net/http"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb/internal"
)

// OptionUserAgent may be passed as an option when creating a client object
// to override the default User-Agent header sent on all requests.
func OptionUserAgent(ua string) kivik.Options {
	return kivik.Options{
		internal.OptionUserAgent: ua,
	}
}

// OptionHTTPClient may be passed as an option when creating a client object,
// to specify an [http.Client].
func OptionHTTPClient(client *http.Client) kivik.Options {
	return kivik.Options{
		internal.OptionHTTPClient: client,
	}
}

// OptionFullCommit is the used to set the `X-Couch-Full-Commit` header in the
// request.
//
// Example:
//
//	db.Put(ctx, "doc_id", doc, couchdb.OptionFullCommit())
func OptionFullCommit() kivik.Options {
	return kivik.Options{
		internal.OptionFullCommit: true,
	}
}

// OptionIfNoneMatch is an option key to set the `If-None-Match header` on
// the request.
//
// Example:
//
//	row, err := db.Get(ctx, "doc_id", couchdb.OptionIfNoneMatch("1-xxx"))
func OptionIfNoneMatch(rev string) kivik.Options {
	return kivik.Options{
		internal.OptionIfNoneMatch: rev,
	}
}

// OptionPartition instructs supporting methods to limit the query to the
// specified partition. Supported methods are: [Query], [AllDocs], [Find],
// and [Explain]. Only supported by CouchDB 3.0.0 and newer.
//
// See https://docs.couchdb.org/en/stable/api/partitioned-dbs.html
func OptionPartition(partition string) kivik.Options {
	return kivik.Options{
		internal.OptionPartition: partition,
	}
}

// OptionNoMultipartPut instructs [github.com/go-kivik/kivik/v4.DB.Put] not
// to use CouchDB's multipart/related upload capabilities. This only affects
// PUT requests that also include attachments.
func OptionNoMultipartPut() kivik.Options {
	return kivik.Options{
		internal.OptionNoMultipartPut: true,
	}
}

// OptionNoMultipartGet instructs [github.com/go-kivik/kivik/v4.DB.Get] not
// to use CouchDB's ability to download attachments with the
// multipart/related media type. This only affects GET requests that request
// attachments.
func OptionNoMultipartGet() kivik.Options {
	return kivik.Options{
		internal.OptionNoMultipartGet: true,
	}
}

// OptionNoCompressedResponses disables gzip content encoding for request
// bodies. Only valid as an option to [github.com/go-kivik/kivik/v4.New].
func OptionNoCompressedResponses() kivik.Options {
	return kivik.Options{
		internal.OptionNoCompressedResponses: true,
	}
}
