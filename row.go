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

import "io"

// Row contains the result of calling Get for a single document. For most uses,
// it is sufficient just to call the ScanDoc method. For more advanced uses, the
// fields may be accessed directly.
type Row struct {
	// ContentLength records the size of the JSON representation of the document
	// as requestd. The value -1 indicates that the length is unknown. Values
	// >= 0 indicate that the given number of bytes may be read from Body.
	ContentLength int64

	// Rev is the revision ID of the returned document.
	Rev string

	// Body represents the document's content.
	//
	// Kivik will always return a non-nil Body, except when Err is non-nil. The
	// ScanDoc method will close Body. When not using ScanDoc, it is the
	// caller's responsibility to close Body
	Body io.ReadCloser

	// Err contains any error that occurred while fetching the document. It is
	// typically returned by ScanDoc.
	Err error

	// Attachments is experimental
	Attachments *AttachmentsIterator
}
