package kivik

import (
	"context"
	"encoding/json"
	"io"
	"testing"

	"github.com/flimzy/kivik/driver"
)

type nonFlusher struct{}

var _ driver.DB = &nonFlusher{}

func (n *nonFlusher) AllDocs(_ context.Context, _ map[string]interface{}) (driver.Rows, error) {
	return nil, nil
}
func (n *nonFlusher) BulkDocs(_ context.Context, _ []interface{}) (driver.BulkResults, error) {
	return nil, nil
}
func (n *nonFlusher) Changes(_ context.Context, _ map[string]interface{}) (driver.Changes, error) {
	return nil, nil
}
func (n *nonFlusher) CreateDoc(_ context.Context, _ interface{}) (string, string, error) {
	return "", "", nil
}
func (n *nonFlusher) DeleteAttachment(_ context.Context, _, _, _ string) (string, error) {
	return "", nil
}
func (n *nonFlusher) Get(_ context.Context, _ string, _ map[string]interface{}) (json.RawMessage, error) {
	return nil, nil
}
func (n *nonFlusher) GetAttachment(_ context.Context, _, _, _ string) (string, driver.MD5sum, io.ReadCloser, error) {
	return "", driver.MD5sum{}, nil, nil
}
func (n *nonFlusher) PutAttachment(_ context.Context, _, _, _, _ string, _ io.Reader) (string, error) {
	return "", nil
}
func (n *nonFlusher) Query(_ context.Context, _, _ string, _ map[string]interface{}) (driver.Rows, error) {
	return nil, nil
}
func (n *nonFlusher) Compact(_ context.Context) error                                { return nil }
func (n *nonFlusher) CompactView(_ context.Context, _ string) error                  { return nil }
func (n *nonFlusher) Delete(_ context.Context, _, _ string) (string, error)          { return "", nil }
func (n *nonFlusher) Put(_ context.Context, _ string, _ interface{}) (string, error) { return "", nil }
func (n *nonFlusher) Security(_ context.Context) (*driver.Security, error)           { return nil, nil }
func (n *nonFlusher) SetSecurity(_ context.Context, _ *driver.Security) error        { return nil }
func (n *nonFlusher) Stats(_ context.Context) (*driver.DBStats, error)               { return nil, nil }
func (n *nonFlusher) ViewCleanup(_ context.Context) error                            { return nil }

func TestFlushNotSupported(t *testing.T) {
	db := &DB{
		driverDB: &nonFlusher{},
	}
	err := db.Flush(context.Background())
	if StatusCode(err) != StatusNotImplemented {
		t.Errorf("Expected NotImplemented, got %s", err)
	}
}
