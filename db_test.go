package kivik

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik/driver"
)

type dummyDB struct{}

var _ driver.DB = &dummyDB{}

func (n *dummyDB) AllDocs(_ context.Context, _ map[string]interface{}) (driver.Rows, error) {
	return nil, nil
}
func (n *dummyDB) BulkDocs(_ context.Context, _ []interface{}) (driver.BulkResults, error) {
	return nil, nil
}
func (n *dummyDB) Changes(_ context.Context, _ map[string]interface{}) (driver.Changes, error) {
	return nil, nil
}
func (n *dummyDB) CreateDoc(_ context.Context, _ interface{}) (string, string, error) {
	return "", "", nil
}
func (n *dummyDB) DeleteAttachment(_ context.Context, _, _, _ string) (string, error) {
	return "", nil
}
func (n *dummyDB) Get(_ context.Context, _ string, _ map[string]interface{}) (json.RawMessage, error) {
	return nil, nil
}
func (n *dummyDB) GetAttachment(_ context.Context, _, _, _ string) (string, driver.MD5sum, io.ReadCloser, error) {
	return "", driver.MD5sum{}, nil, nil
}
func (n *dummyDB) PutAttachment(_ context.Context, _, _, _, _ string, _ io.Reader) (string, error) {
	return "", nil
}
func (n *dummyDB) Query(_ context.Context, _, _ string, _ map[string]interface{}) (driver.Rows, error) {
	return nil, nil
}
func (n *dummyDB) Compact(_ context.Context) error                                { return nil }
func (n *dummyDB) CompactView(_ context.Context, _ string) error                  { return nil }
func (n *dummyDB) Delete(_ context.Context, _, _ string) (string, error)          { return "", nil }
func (n *dummyDB) Put(_ context.Context, _ string, _ interface{}) (string, error) { return "", nil }
func (n *dummyDB) Security(_ context.Context) (*driver.Security, error)           { return nil, nil }
func (n *dummyDB) SetSecurity(_ context.Context, _ *driver.Security) error        { return nil }
func (n *dummyDB) Stats(_ context.Context) (*driver.DBStats, error)               { return nil, nil }
func (n *dummyDB) ViewCleanup(_ context.Context) error                            { return nil }

func TestFlushNotSupported(t *testing.T) {
	db := &DB{
		driverDB: &dummyDB{},
	}
	err := db.Flush(context.Background())
	if StatusCode(err) != StatusNotImplemented {
		t.Errorf("Expected NotImplemented, got %s", err)
	}
}

type putGrabber struct {
	*dummyDB
	lastPut interface{}
}

func (db *putGrabber) Put(_ context.Context, _ string, i interface{}) (string, error) {
	db.lastPut = i
	return "", nil
}

type errorReader struct{}

var _ io.Reader = &errorReader{}

func (r *errorReader) Read(_ []byte) (int, error) {
	return 0, errors.New("errorReader")
}

func TestNormalizeFromJSON(t *testing.T) {
	type njTest struct {
		Name     string
		Input    interface{}
		Expected interface{}
		Status   int
		Error    string
	}
	tests := []njTest{
		{
			Name:     "Interface",
			Input:    int(5),
			Expected: int(5),
		},
		{
			Name:   "InvalidJSON",
			Input:  []byte(`invalid`),
			Status: 400,
			Error:  "invalid character 'i' looking for beginning of value",
		},
		{
			Name:     "Bytes",
			Input:    []byte(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:     "RawMessage",
			Input:    json.RawMessage(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:     "ioReader",
			Input:    strings.NewReader(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:   "ErrorReader",
			Input:  &errorReader{},
			Status: 500,
			Error:  "errorReader",
		},
	}
	for _, test := range tests {
		func(test njTest) {
			t.Run(test.Name, func(t *testing.T) {
				result, err := normalizeFromJSON(test.Input)
				var msg string
				var status int
				if err != nil {
					msg = err.Error()
					status = StatusCode(err)
				}
				if msg != test.Error || status != test.Status {
					t.Errorf("Unexpected error: %d %s", status, msg)
				}
				if err != nil {
					return
				}
				if d := diff.Interface(test.Expected, result); d != "" {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestPutJSON(t *testing.T) {
	grabber := &putGrabber{}
	db := &DB{driverDB: grabber}
	type putTest struct {
		Name     string
		Input    interface{}
		Status   int
		Expected interface{}
		Error    string
	}
	tests := []putTest{
		{
			Name:     "Interface",
			Input:    map[string]string{"foo": "bar"},
			Expected: map[string]string{"foo": "bar"},
		},
		{
			Name:   "InvalidJSON",
			Input:  []byte("Something bogus"),
			Status: 400,
			Error:  "invalid character 'S' looking for beginning of value",
		},
		{
			Name:     "Bytes",
			Input:    []byte(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:     "RawMessage",
			Input:    json.RawMessage(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:     "Reader",
			Input:    strings.NewReader(`{"foo":"bar"}`),
			Expected: map[string]interface{}{"foo": "bar"},
		},
		{
			Name:   "ErrorReader",
			Input:  &errorReader{},
			Status: 500,
			Error:  "errorReader",
		},
	}
	for _, test := range tests {
		func(test putTest) {
			t.Run(test.Name, func(t *testing.T) {
				_, err := db.Put(context.Background(), "foo", test.Input)
				var msg string
				var status int
				if err != nil {
					msg = err.Error()
					status = StatusCode(err)
				}
				if msg != test.Error || status != test.Status {
					t.Errorf("Unexpected error: %d %s", status, msg)
				}
				if err != nil {
					return
				}
				if d := diff.Interface(test.Expected, grabber.lastPut); d != "" {
					t.Error(d)
				}
			})
		}(test)
	}
}
