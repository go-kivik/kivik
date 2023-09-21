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

package cdb

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"path/filepath"

	"github.com/go-kivik/fsdb/v4/filesystem"
	"github.com/go-kivik/kivik/v4/driver"
)

/*
When uploading attachment stubs:
- revpos must match, or be omitted
- digest must match
- length is ignored
- content_type is ignored
*/

// Attachment represents a file attachment.
type Attachment struct {
	ContentType string `json:"content_type" yaml:"content_type"`
	RevPos      *int64 `json:"revpos,omitempty" yaml:"revpos,omitempty"`
	Stub        bool   `json:"stub,omitempty" yaml:"stub,omitempty"`
	Follows     bool   `json:"follows,omitempty" yaml:"follows,omitempty"`
	Content     []byte `json:"data,omitempty" yaml:"content,omitempty"`
	Size        int64  `json:"length" yaml:"size"`
	Digest      string `json:"digest" yaml:"digest"`

	// path is the full path to the file on disk, or the empty string if the
	// attachment is not (yet) on disk.
	path string
	// fs is the filesystem to use for disk access.
	fs filesystem.Filesystem

	// outputStub dictates whether MarshalJSON should output a stub. This is
	// distinct from Stub, which indicates whether UnmarshalJSON read Stub, as
	// from user input.
	outputStub bool
}

// Open opens the attachment for reading.
func (a *Attachment) Open() (filesystem.File, error) {
	if a.path == "" {
		return nil, errors.New("no path defined")
	}
	return a.fs.Open(a.path)
}

// MarshalJSON implements the json.Marshaler interface.
func (a *Attachment) MarshalJSON() ([]byte, error) {
	var err error
	switch {
	case len(a.Content) != 0:
		a.setMetadata()
	case a.outputStub || a.Follows:
		err = a.readMetadata()
	default:
		err = a.readContent()
	}
	if err != nil {
		return nil, err
	}
	att := struct {
		Attachment
		Content *[]byte `json:"data,omitempty"`    // nolint: govet
		Stub    *bool   `json:"stub,omitempty"`    // nolint: govet
		Follows *bool   `json:"follows,omitempty"` // nolint: govet
	}{
		Attachment: *a,
	}
	switch {
	case a.outputStub:
		att.Stub = &a.outputStub
	case a.Follows:
		att.Follows = &a.Follows
	case len(a.Content) > 0:
		att.Content = &a.Content
	}
	return json.Marshal(att)
}

func (a *Attachment) readContent() error {
	f, err := a.fs.Open(a.path)
	if err != nil {
		return err
	}
	buf := &bytes.Buffer{}
	a.Size, a.Digest, err = copyDigest(buf, f)
	if err != nil {
		return err
	}
	a.Content = buf.Bytes()
	return nil
}

func (a *Attachment) readMetadata() error {
	if a.path == "" {
		return nil
	}
	f, err := a.fs.Open(a.path)
	if err != nil {
		return err
	}
	a.Size, a.Digest = digest(f)
	return nil
}

func (a *Attachment) setMetadata() {
	a.Size, a.Digest = digest(bytes.NewReader(a.Content))
}

func (a *Attachment) persist(path, attname string) error {
	target := filepath.Join(path, attname)
	if err := atomicWriteFile(a.fs, target, bytes.NewReader(a.Content)); err != nil {
		return err
	}
	a.Content = nil
	a.path = target
	return nil
}

type attsIter []*driver.Attachment

var _ driver.Attachments = &attsIter{}

func (i attsIter) Close() error {
	for _, att := range i {
		if err := att.Content.Close(); err != nil {
			return err
		}
	}
	return nil
}

func (i *attsIter) Next(att *driver.Attachment) error {
	if len(*i) == 0 {
		return io.EOF
	}
	var next *driver.Attachment
	next, *i = (*i)[0], (*i)[1:]
	*att = *next
	return nil
}

// AttachmentsIterator will return a driver.Attachments iterator, if the options
// permit. If options don't permit, both return values will be nil.
func (r *Revision) AttachmentsIterator() (driver.Attachments, error) {
	if attachments, _ := r.options["attachments"].(bool); !attachments {
		return nil, nil
	}
	if accept, _ := r.options["header:accept"].(string); accept == "application/json" {
		return nil, nil
	}
	iter := make(attsIter, 0, len(r.Attachments))
	for filename, att := range r.Attachments {
		f, err := att.Open()
		if err != nil {
			return nil, err
		}
		drAtt := &driver.Attachment{
			Filename:    filename,
			Content:     f,
			ContentType: att.ContentType,
			Stub:        att.Stub,
			Follows:     att.Follows,
			Size:        att.Size,
			Digest:      att.Digest,
		}
		if att.RevPos != nil {
			drAtt.RevPos = *att.RevPos
		}
		iter = append(iter, drAtt)
	}
	return &iter, nil
}
