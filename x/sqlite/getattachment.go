package sqlite

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) GetAttachment(ctx context.Context, docID string, filename string, _ driver.Options) (*driver.Attachment, error) {
	var found bool

	err := d.db.QueryRowContext(ctx, fmt.Sprintf(`
		SELECT EXISTS (
			SELECT 1
			FROM %[1]q AS att
			WHERE att.id = $1
			AND att.filename = $2
		)
		`, d.name+"_attachments"), docID, filename).Scan(&found)
	if err != nil {
		return nil, err
	}
	if found {
		return &driver.Attachment{}, nil
	}
	return nil, &internal.Error{Status: http.StatusNotFound, Message: "Not Found: missing"}
}
