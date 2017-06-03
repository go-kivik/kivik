package kivik

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/flimzy/kivik/driver"
)

type rows struct{}

var _ driver.Rows = &rows{}

func (r *rows) Close() error             { return nil }
func (r *rows) Next(_ *driver.Row) error { return nil }
func (r *rows) Offset() int64            { return 0 }
func (r *rows) TotalRows() int64         { return 0 }
func (r *rows) UpdateSeq() string        { return "" }

type wrows struct {
	*rows
}

var _ driver.RowsWarner = &wrows{}

func (r *wrows) Warning() string { return "test warning" }

func TestWarning(t *testing.T) {
	t.Run("Warner", func(t *testing.T) {
		r := newRows(context.Background(), &wrows{})
		expected := "test warning"
		if w := r.Warning(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})
	t.Run("NonWarner", func(t *testing.T) {
		r := newRows(context.Background(), &rows{})
		expected := ""
		if w := r.Warning(); w != expected {
			t.Errorf("Warning\nExpected: %s\n  Actual: %s", expected, w)
		}
	})
}
