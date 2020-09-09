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

package kivik

const (
	// KivikVersion is the version of the Kivik library.
	KivikVersion = "4.0.0-prerelease"
	// KivikVendor is the vendor string reported by this library.
	KivikVendor = "Kivik"
)

// SessionCookieName is the name of the CouchDB session cookie.
const SessionCookieName = "AuthSession"

// UserPrefix is the mandatory CouchDB user prefix.
// See http://docs.couchdb.org/en/2.0.0/intro/security.html#org-couchdb-user
const UserPrefix = "org.couchdb.user:"

// EndKeySuffix is a high Unicode character (0xfff0) useful for appending to an
// endkey argument, when doing a ranged search, as described here:
// http://couchdb.readthedocs.io/en/latest/ddocs/views/collation.html#string-ranges
//
// Example, to return all results with keys beginning with "foo":
//
//    rows, err := db.Query(context.TODO(), "ddoc", "view", map[string]interface{}{
//        "startkey": "foo",
//        "endkey":   "foo" + kivik.EndKeySuffix,
//    })
const EndKeySuffix = string(rune(0xfff0))
