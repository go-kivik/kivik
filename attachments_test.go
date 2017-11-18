package kivik

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
)

func TestAttachmentBytes(t *testing.T) {
	tests := []struct {
		name     string
		att      *Attachment
		expected string
		err      string
	}{
		{
			name:     "read success",
			att:      NewAttachment("test.txt", "text/plain", ioutil.NopCloser(strings.NewReader("test content"))),
			expected: "test content",
		},
		{
			name: "buffered read",
			att: func() *Attachment {
				att := NewAttachment("test.txt", "text/plain", ioutil.NopCloser(strings.NewReader("test content")))
				_, _ = att.Bytes()
				return att
			}(),
			expected: "test content",
		},
		{
			name: "read error",
			att:  NewAttachment("test.txt", "text/plain", errReader("read error")),
			err:  "read error",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.att.Bytes()
			testy.Error(t, test.err, err)
			if d := diff.Text(test.expected, string(result)); d != nil {
				t.Error(d)
			}
		})
	}
}
