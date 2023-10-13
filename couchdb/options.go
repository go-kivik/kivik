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
	"fmt"
	"net/http"
	"path"
	"strings"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb/chttp"
)

type optionHTTPClient struct {
	*http.Client
}

func (c optionHTTPClient) Apply(target interface{}) {
	if client, ok := target.(*http.Client); ok {
		*client = *c.Client
	}
}

func (optionHTTPClient) String() string { return "custom *http.Client" }

// OptionHTTPClient may be passed as an option when creating a CouchDB client
// to specify an custom [net/http.Client] to be used when making all API calls.
// Only honored by [github.com/go-kivik/kivik/v4.New].
func OptionHTTPClient(client *http.Client) kivik.Option {
	return optionHTTPClient{Client: client}
}

// OptionNoRequestCompression instructs the CouchDB client not to use gzip
// compression for request bodies sent to the server. Only honored by
// [github.com/go-kivik/kivik/v4.New].
func OptionNoRequestCompression() kivik.Option {
	return chttp.OptionNoRequestCompression()
}

// OptionUserAgent may be passed as an option when creating a client object,
// to append to the default User-Agent header sent on all requests.
func OptionUserAgent(ua string) kivik.Option {
	return chttp.OptionUserAgent(ua)
}

// OptionFullCommit is the option key used to set the `X-Couch-Full-Commit`
// header in the request when set to true.
func OptionFullCommit() kivik.Option {
	return chttp.OptionFullCommit()
}

// OptionIfNoneMatch is an option key to set the `If-None-Match` header on
// the request.
func OptionIfNoneMatch(value string) kivik.Option {
	return chttp.OptionIfNoneMatch(value)
}

type partitionedPath struct {
	path string
	part string
}

func partPath(path string) *partitionedPath {
	return &partitionedPath{
		path: path,
	}
}

func (pp partitionedPath) String() string {
	if pp.part == "" {
		return pp.path
	}
	return path.Join("_partition", pp.part, pp.path)
}

type optionPartition string

func (o optionPartition) Apply(target interface{}) {
	if ppath, ok := target.(*partitionedPath); ok {
		ppath.part = string(o)
	}
}

func (o optionPartition) String() string {
	return fmt.Sprintf("[partition:%s]", string(o))
}

// OptionPartition instructs supporting methods to limit the query to the
// specified partition. Supported methods are: Query, AllDocs, Find, and
// Explain. Only supported by CouchDB 3.0.0 and newer.
//
// See the [CouchDB documentation].
//
// [CouchDB documentation]: https://docs.couchdb.org/en/stable/api/partitioned-dbs.html
func OptionPartition(partition string) kivik.Option {
	return optionPartition(partition)
}

type optionNoMultipartPut struct{}

func (optionNoMultipartPut) Apply(target interface{}) {
	if putOpts, ok := target.(*putOptions); ok {
		putOpts.NoMultipartPut = true
	}
}

func (optionNoMultipartPut) String() string {
	return "[NoMultipartPut]"
}

// OptionNoMultipartPut instructs [github.com/go-kivik/kivik/v4.DB.Put] not
// to use CouchDB's multipart/related upload capabilities. This is only honored
// by calls to [github.com/go-kivik/kivik/v4.DB.Put] that also include
// attachments.
func OptionNoMultipartPut() kivik.Option {
	return optionNoMultipartPut{}
}

type optionNoMultipartGet struct{}

func (optionNoMultipartGet) Apply(target interface{}) {
	if getOpts, ok := target.(*getOptions); ok {
		getOpts.noMultipartGet = true
	}
}

func (optionNoMultipartGet) String() string {
	return "[NoMultipartGet]"
}

// OptionNoMultipartGet instructs [github.com/go-kivik/kivik/v4.DB.Get] not
// to use CouchDB's ability to download attachments with the
// multipart/related media type. This is only honored by calls to
// [github.com/go-kivik/kivik/v4.DB.Get] that request attachments.
func OptionNoMultipartGet() kivik.Option {
	return optionNoMultipartGet{}
}

type multiOptions []kivik.Option

var _ kivik.Option = (multiOptions)(nil)

func (o multiOptions) Apply(t interface{}) {
	for _, opt := range o {
		if opt != nil {
			opt.Apply(t)
		}
	}
}

func (o multiOptions) String() string {
	parts := make([]string, 0, len(o))
	for _, opt := range o {
		if part := fmt.Sprintf("%s", opt); part != "" {
			parts = append(parts, part)
		}
	}
	return strings.Join(parts, ",")
}
