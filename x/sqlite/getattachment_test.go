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
	type attachmentMetadata struct {
		Filename    string
		ContentType string
		Length      int64
		RevPos      int64
	}
	type test struct {
		setup    func(t *testing.T, db driver.DB)
		docID    string
		filename string

		wantAttachment *attachmentMetadata
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
	tests.Add("when the attachment exists, return it", test{
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
	tests.Add("when an attachment is returned, it contains metadata...", test{
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
		wantAttachment: &attachmentMetadata{
			Filename:    "foo.txt",
			ContentType: "text/plain",
			Length:      25,
			RevPos:      1,
		},
	})
	// GetAttachment returns the latest revision by default
	//

	/*
		TODO:
		- doc exists, and file exists, but doc is deleted
		- return correct attachment in case of a conflict
		- return existing file from existing doc
		- request attachment from historical revision
		- failure: request attachment from historical revision that does not exist



		- GetAttachment returns 404 when the document does exist, but the attachment has never existed
		- GetAttachment returns 404 when the document has never existed
		- GetAttachment returns 404 when the document was deleted
		- GetAttachment returns 404 when the latest revision was deleted
		- GetAttachment returns 404 when the document does exist, but the attachment has been deleted
		- GetAttachment returns the latest revision
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
		attachment, err := db.GetAttachment(context.Background(), tt.docID, tt.filename, opts)
		if !testy.ErrorMatches(tt.wantErr, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if status := kivik.HTTPStatus(err); status != tt.wantStatus {
			t.Errorf("Unexpected status: %d", status)
		}

		if tt.wantAttachment == nil {
			return
		}
		got := &attachmentMetadata{
			Filename:    attachment.Filename,
			ContentType: attachment.ContentType,
			Length:      attachment.Size,
			RevPos:      attachment.RevPos,
		}
		if d := cmp.Diff(tt.wantAttachment, got); d != "" {
			t.Errorf("Unexpected attachment metadata:\n%s", d)
		}
	})
}
