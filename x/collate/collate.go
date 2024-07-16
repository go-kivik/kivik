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

// Package collate provides (near) CouchDB-compatible collation functions.
//
// The collation order provided by this package differs slightly from that
// described by the [CouchDB documentation]. In particular:
//
//   - The Unicode UCI algorithm supported natively by Go sorts the backtick (`)
//     and caret (^) after other symbols, not before.
//   - Because Go's maps are unordered, this implementation does not honor the
//     order of object key members when collating.  That is to say, the object
//     `{b:2,a:1}` is treated as equivalent to `{a:1,b:2}` for collation
//     purposes. This is tracked in [issue #952]. Please leave a comment there
//     if this is affecting you.
//
// [CouchDB documentation]: https://docs.couchdb.org/en/stable/ddocs/views/collation.html#collation-specification
// [issue #952]: https://github.com/go-kivik/kivik/issues/952
package collate

import (
	"sync"

	"golang.org/x/text/collate"
	"golang.org/x/text/language"
)

var (
	collatorMu = new(sync.Mutex)
	collator   = collate.New(language.Und)
)

// CompareString returns an integer comparing the two strings.
// The result will be 0 if a==b, -1 if a < b, and +1 if a > b.
func CompareString(a, b string) int {
	collatorMu.Lock()
	defer collatorMu.Unlock()
	return collator.CompareString(a, b)
}
