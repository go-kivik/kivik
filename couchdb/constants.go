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

import "github.com/go-kivik/kivik/v4/couchdb/internal"

// Version is the current version of this package.
const Version = "4.0.0-prerelease"

const (
	// OptionNoMultipartPut instructs [github.com/go-kivik/kivik/v4.DB.Put] not
	// to use CouchDB's multipart/related upload capabilities. This only affects
	// PUT requests that also include attachments.
	OptionNoMultipartPut = internal.OptionNoMultipartPut

	// OptionNoMultipartGet instructs [github.com/go-kivik/kivik/v4.DB.Get] not
	// to use CouchDB's ability to download attachments with the
	// multipart/related media type. This only affects GET requests that request
	// attachments.
	OptionNoMultipartGet = internal.OptionNoMultipartGet
)

const (
	typeJSON      = "application/json"
	typeMPRelated = "multipart/related"
)
