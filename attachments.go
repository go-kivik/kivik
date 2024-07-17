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

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/go-kivik/kivik/v4/driver"
)

// Attachments is a collection of one or more file attachments.
type Attachments map[string]*Attachment

// Get fetches the requested attachment, or returns nil if it does not exist.
func (a *Attachments) Get(filename string) *Attachment {
	return map[string]*Attachment(*a)[filename]
}

// Set sets the attachment associated with filename in the collection,
// replacing it if it already exists.
func (a *Attachments) Set(filename string, att *Attachment) {
	map[string]*Attachment(*a)[filename] = att
}

// Delete removes the specified file from the collection.
func (a *Attachments) Delete(filename string) {
	delete(map[string]*Attachment(*a), filename)
}

// Attachment represents a file attachment on a CouchDB document.
type Attachment struct {
	// Filename is the name of the attachment.
	Filename string `json:"-"`

	// ContentType is the Content-Type type of the attachment.
	ContentType string `json:"content_type"`

	// Stub will be true if the data structure only represents file metadata,
	// and contains no actual content. Stub will be true when returned by
	// [DB.GetAttachmentMeta], or when included in a document without the
	// 'include_docs' option.
	Stub bool `json:"stub"`

	// Follows will be true when reading attachments in multipart/related
	// format.
	Follows bool `json:"follows"`

	// Content represents the attachment's content.
	//
	// Kivik will always return a non-nil Content, even for 0-byte attachments
	// or when Stub is true. It is the caller's responsibility to close
	// Content.
	Content io.ReadCloser `json:"-"`

	// Size records the uncompressed size of the attachment. The value -1
	// indicates that the length is unknown. Unless [Attachment.Stub] is true,
	// values >= 0 indicate that the given number of bytes may be read from
	// [Attachment.Content].
	Size int64 `json:"length"`

	// Used compression codec, if any. Will be the empty string if the
	// attachment is uncompressed.
	ContentEncoding string `json:"encoding"`

	// EncodedLength records the compressed attachment size in bytes. Only
	// meaningful when [Attachment.ContentEncoding] is defined.
	EncodedLength int64 `json:"encoded_length"`

	// RevPos is the revision number when attachment was added.
	RevPos int64 `json:"revpos"`

	// Digest is the content hash digest.
	Digest string `json:"digest"`
}

// bufCloser wraps a *bytes.Buffer to create an io.ReadCloser
type bufCloser struct {
	*bytes.Buffer
}

var _ io.ReadCloser = &bufCloser{}

func (b *bufCloser) Close() error { return nil }

// validate returns an error if the attachment is invalid.
func (a *Attachment) validate() error {
	if a == nil {
		return missingArg("attachment")
	}
	if a.Filename == "" {
		return missingArg("filename")
	}
	return nil
}

// MarshalJSON satisfies the [encoding/json.Marshaler] interface.
func (a *Attachment) MarshalJSON() ([]byte, error) {
	type jsonAttachment struct {
		ContentType string `json:"content_type"`
		Stub        *bool  `json:"stub,omitempty"`
		Follows     *bool  `json:"follows,omitempty"`
		Size        int64  `json:"length,omitempty"`
		RevPos      int64  `json:"revpos,omitempty"`
		Data        []byte `json:"data,omitempty"`
		Digest      string `json:"digest,omitempty"`
	}
	att := &jsonAttachment{
		ContentType: a.ContentType,
		Size:        a.Size,
		RevPos:      a.RevPos,
		Digest:      a.Digest,
	}
	switch {
	case a.Stub:
		att.Stub = &a.Stub
	case a.Follows:
		att.Follows = &a.Follows
	default:
		defer a.Content.Close() // nolint: errcheck
		data, err := io.ReadAll(a.Content)
		if err != nil {
			return nil, err
		}
		att.Data = data
	}
	return json.Marshal(att)
}

// UnmarshalJSON implements the [encoding/json.Unmarshaler] interface.
func (a *Attachment) UnmarshalJSON(data []byte) error {
	type clone Attachment
	type jsonAtt struct {
		clone
		Data []byte `json:"data"`
	}
	var att jsonAtt
	if err := json.Unmarshal(data, &att); err != nil {
		return err
	}
	*a = Attachment(att.clone)
	if att.Data != nil {
		a.Content = io.NopCloser(bytes.NewReader(att.Data))
	} else {
		a.Content = nilContent
	}
	return nil
}

// UnmarshalJSON implements the [encoding/json.Unmarshaler] interface.
func (a *Attachments) UnmarshalJSON(data []byte) error {
	atts := make(map[string]*Attachment)
	if err := json.Unmarshal(data, &atts); err != nil {
		return err
	}
	for filename, att := range atts {
		att.Filename = filename
	}
	*a = atts
	return nil
}

// AttachmentsIterator allows reading streamed attachments from a multi-part
// [DB.Get] request.
type AttachmentsIterator struct {
	atti    driver.Attachments
	onClose func()
}

// Next returns the next attachment in the stream. [io.EOF] will be
// returned when there are no more attachments.
//
// The returned attachment is only valid until the next call to [Next], or a
// call to [Close].
func (i *AttachmentsIterator) Next() (*Attachment, error) {
	att := new(driver.Attachment)
	if err := i.atti.Next(att); err != nil {
		if err == io.EOF {
			if e2 := i.Close(); e2 != nil {
				return nil, e2
			}
		}
		return nil, err
	}
	katt := Attachment(*att)
	return &katt, nil
}

// Close closes the AttachmentsIterator. It is automatically called when
// [AttachmentsIterator.Next] returns [io.EOF].
func (i *AttachmentsIterator) Close() error {
	if i.onClose != nil {
		i.onClose()
	}
	return i.atti.Close()
}

// Iterator returns a function that can be used to iterate over the attachments.
// This function works with Go 1.23's range functions, and is an alternative to
// using [AttachmentsIterator.Next] directly.
func (i *AttachmentsIterator) Iterator() func(yield func(*Attachment, error) bool) {
	return func(yield func(*Attachment, error) bool) {
		for {
			att, err := i.Next()
			if err == io.EOF {
				return
			}
			if !yield(att, err) || err != nil {
				_ = i.Close()
				return
			}
		}
	}
}
