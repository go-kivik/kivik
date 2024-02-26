package sqlite

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBGetAttachment(t *testing.T) {
	t.Parallel()
	type test struct {
		docID    string
		filename string

		wantAttachment *driver.Attachment
		wantStatus     int
		wantErr        string
	}

	tests := testy.NewTable()
	tests.Add("document does not exist", test{
		docID:      "foo",
		filename:   "foo.txt",
		wantStatus: http.StatusNotFound,
		wantErr:    "Not Found: missing",
	})
	/*
		TODO:
		- doc exists, but filename does not
		- doc exists, and file exists, but doc is deleted
		- return correct attachment in case of a conflict
		- return existing file from existing doc
		- request attachment from historical revision
		- failure: request attachment from historical revision that does not exist
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := newDB(t)
		// if tt.setup != nil {
		// 	tt.setup(t, db)
		// }
		// opts := tt.options
		// if opts == nil {
		opts := mock.NilOption
		// }
		attachment, err := db.GetAttachment(context.Background(), tt.docID, tt.filename, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
		if err != nil {
			return
		}
		if d := cmp.Diff(tt.wantAttachment, attachment); d != "" {
			t.Errorf("Unexpected attachment:\n%s\n", d)
		}
	})
}
