package sqlite

import (
	"context"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestDBGetAttachment(t *testing.T) {
	t.Parallel()
	type test struct {
		setup    func(t *testing.T, db driver.DB)
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
	tests.Add("when document exists and file exists we get the valid status", test{
		setup: func(t *testing.T, db driver.DB) {
			_, err := db.Put(context.Background(), "foo", map[string]interface{}{
				"_id": "foo",
				"_attachments": map[string]interface{}{
					"foo.txt": map[string]interface{}{
						"content_type": "text/plain",
						"data":         "VGhpcyBpcyBhIGJhc2U2NCBlbmNvZGluZw==",
					},
				},
			}, mock.NilOption)
			if err != nil {
				t.Fatal(err)
			}
		},
		docID:    "foo",
		filename: "foo.txt",
	})
	/*
		TODO:
		- doc exists, and file exists, but doc is deleted
		- return correct attachment in case of a conflict
		- return existing file from existing doc
		- request attachment from historical revision
		- failure: request attachment from historical revision that does not exist
	*/

	tests.Run(t, func(t *testing.T, tt test) {
		t.Parallel()
		db := newDB(t)
		if tt.setup != nil {
			tt.setup(t, db)
		}
		// opts := tt.options
		// if opts == nil {
		opts := mock.NilOption
		// }
		_, err := db.GetAttachment(context.Background(), tt.docID, tt.filename, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}
	})
}
