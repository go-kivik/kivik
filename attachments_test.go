package kivik

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestAttachmentBytes(t *testing.T) {
	content := "test content"
	att := NewAttachment("test.txt", "text/plain", ioutil.NopCloser(strings.NewReader(content)))
	result, err := att.Bytes()
	if err != nil {
		t.Fatalf("read failed: %s", err)
	}
	if string(result) != content {
		t.Errorf("Read unexpected.\nExpected: %s\n  Actual: %s\n", content, result)
	}

	result2, err := att.Bytes()
	if err != nil {
		t.Fatalf("second read failed: %s", err)
	}
	if string(result2) != content {
		t.Errorf("Second read unexpected.\nExpected: %s\n  Actual: %s\n", content, result2)
	}
}
