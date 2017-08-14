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

func TestClient(t *testing.T) {
	client := &Client{}
	db := &DB{client: client}
	result := db.Client()
	if result != client {
		t.Errorf("Unexpected result. Expected %p, got %p", client, result)
	}
}

func TestName(t *testing.T) {
	dbName := "foo"
	db := &DB{name: dbName}
	result := db.Name()
	if result != dbName {
		t.Errorf("Unexpected result. Expected %s, got %s", dbName, result)
	}
}

type dummyDB struct {
	driver.DB
}

var _ driver.DB = &dummyDB{}

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
				if d := diff.Interface(test.Expected, result); d != nil {
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
				if d := diff.Interface(test.Expected, grabber.lastPut); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestExtractDocID(t *testing.T) {
	type ediTest struct {
		name     string
		i        interface{}
		id       string
		expected bool
	}
	tests := []ediTest{
		{
			name: "nil",
		},
		{
			name: "string/interface map, no id",
			i: map[string]interface{}{
				"value": "foo",
			},
		},
		{
			name: "string/interface map, with id",
			i: map[string]interface{}{
				"_id": "foo",
			},
			id:       "foo",
			expected: true,
		},
		{
			name: "string/string map, with id",
			i: map[string]string{
				"_id": "foo",
			},
			id:       "foo",
			expected: true,
		},
		{
			name: "invalid JSON",
			i:    make(chan int),
		},
		{
			name: "valid JSON",
			i: struct {
				ID string `json:"_id"`
			}{ID: "oink"},
			id:       "oink",
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			id, ok := extractDocID(test.i)
			if ok != test.expected || test.id != id {
				t.Errorf("Expected %t/%s, got %t/%s", test.expected, test.id, ok, id)
			}
		})
	}
}
