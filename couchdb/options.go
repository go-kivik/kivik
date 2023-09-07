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
// to specify an custom *http.Client to be used when making all API calls.
func OptionHTTPClient(client *http.Client) kivik.Option {
	return optionHTTPClient{Client: client}
}

// OptionNoRequestCompression instructs the CouchDB client not to use gzip
// compression for request bodies sent to the server. Only honored when passed
// to [github.com/go-kivik/kivik/v4.New].
func OptionNoRequestCompression() kivik.Option {
	return chttp.OptionNoRequestCompression()
}
