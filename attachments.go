package kivik

import (
	"bytes"
	"io"
)

// Attachment represents a CouchDB file attachment.
type Attachment struct {
	io.Reader
	Filename    string
	ContentType string
	MD5         [16]byte
}

// Bytes returns the attachment's body as a byte slice.
func (a *Attachment) Bytes() ([]byte, error) {
	if buf, ok := a.Reader.(*bytes.Buffer); ok {
		// Simple optimization
		return buf.Bytes(), nil
	}
	buf := &bytes.Buffer{}
	if _, err := io.Copy(buf, a); err != nil {
		return nil, err
	}
	a.Reader = buf
	return buf.Bytes(), nil
}

// NewAttachment returns a new CouchDB attachment.
func NewAttachment(filename, contentType string, body io.Reader) *Attachment {
	return &Attachment{
		Reader:      body,
		Filename:    filename,
		ContentType: contentType,
	}
}
