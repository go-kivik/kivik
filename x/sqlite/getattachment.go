package sqlite

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal"
)

func (d *db) GetAttachment(_ context.Context, docID string, filename string, _ driver.Options) (*driver.Attachment, error) {
	if true {
		return nil, &internal.Error{Status: http.StatusNotFound, Message: "Not Found: missing"}
	}
	panic("not implemented")
}
