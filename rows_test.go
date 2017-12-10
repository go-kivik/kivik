package kivik

import (
	"context"
	"testing"
)

func TestWarning(t *testing.T) {
	t.Run("Warner", func(t *testing.T) {
		expected := "test warning"
		r := newRows(context.Background(), &mockRowsWarner{
			WarningFunc: func() string { return expected },
		})
		if w := r.Warning(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})
	t.Run("NonWarner", func(t *testing.T) {
		r := newRows(context.Background(), &mockRows{})
		expected := ""
		if w := r.Warning(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})
}

func TestBookmark(t *testing.T) {
	t.Run("Bookmarker", func(t *testing.T) {
		expected := "test bookmark"
		r := newRows(context.Background(), &mockBookmarker{
			BookmarkFunc: func() string { return expected },
		})
		if w := r.Bookmark(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})
	t.Run("Non Bookmarker", func(t *testing.T) {
		r := newRows(context.Background(), &mockRows{})
		expected := ""
		if w := r.Bookmark(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})
}
