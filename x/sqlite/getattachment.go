package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) GetAttachment(ctx context.Context, docID string, filename string, _ driver.Options) (*driver.Attachment, error) {
	attachment, err := d.attachmentExists(ctx, docID, filename)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "Not Found: missing"}
	}
	if err != nil {
		return nil, err
	}

	return attachment, nil
}

func (d *db) attachmentExists(ctx context.Context, docID string, filename string) (*driver.Attachment, error) {
	var att driver.Attachment
	err := d.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT filename, content_type, length, rev
		FROM %s
		WHERE id = $1 AND filename = $2
		`, d.name+"_attachments"), docID, filename).Scan(&att.Filename, &att.ContentType, &att.Size, &att.RevPos)
	return &att, err
}
