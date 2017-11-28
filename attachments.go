package kivik

import (
	"bytes"
	"io"

	"github.com/flimzy/kivik/errors"
)

// MD5sum is a 128-bit MD5 checksum.
type MD5sum [16]byte

// Attachment represents a file attachment on a CouchDB document.
type Attachment struct {
	io.ReadCloser
	Filename    string
	ContentType string
	MD5         [16]byte
}

var _ io.ReadCloser = Attachment{}

func (a Attachment) Read(p []byte) (int, error) {
	if a.ReadCloser == nil {
		// TODO: Consider an alternative error code for this case
		return 0, errors.Status(StatusUnknownError, "kivik: attachment content not read")
	}
	return a.ReadCloser.Read(p)
}

// Close calls the underlying close method.
func (a Attachment) Close() error {
	if a.ReadCloser == nil {
		return nil
	}
	return a.ReadCloser.Close()
}

// bufCloser wraps a *bytes.Buffer to create an io.ReadCloser
type bufCloser struct {
	*bytes.Buffer
}

var _ io.ReadCloser = &bufCloser{}

func (b *bufCloser) Close() error { return nil }

// Bytes returns the attachment's body as a byte slice.
func (a *Attachment) Bytes() ([]byte, error) {
	if buf, ok := a.ReadCloser.(*bufCloser); ok {
		// Simple optimization
		return buf.Bytes(), nil
	}
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, a); err != nil {
		return nil, err
	}
	a.ReadCloser = &bufCloser{buf}
	return buf.Bytes(), nil
}

// NewAttachment returns a new CouchDB attachment.
func NewAttachment(filename, contentType string, body io.ReadCloser) *Attachment {
	return &Attachment{
		ReadCloser:  body,
		Filename:    filename,
		ContentType: contentType,
	}
}

// validate returns an error if the attachment is invalid.
func (a *Attachment) validate() error {
	if a.Filename == "" {
		return missingArg("filename")
	}
	return nil
}
