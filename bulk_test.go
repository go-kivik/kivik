package kivik

import (
	"context"
	"errors"
	"fmt"
	"io"
	"reflect"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik/driver"
)

func TestDocsInterfaceSlice(t *testing.T) {
	type diTest struct {
		name     string
		input    interface{}
		expected interface{}
		error    string
	}
	str := "foo"
	intSlice := []int{1, 2, 3}
	tests := []diTest{
		{
			name:     "Nil",
			input:    nil,
			expected: nil,
			error:    "must be slice or array, got <nil>",
		},
		{
			name:     "InterfaceSlice",
			input:    []interface{}{map[string]string{"foo": "bar"}},
			expected: []interface{}{map[string]string{"foo": "bar"}},
		},
		{
			name:  "String",
			input: "foo",
			error: "must be slice or array, got string",
		},
		{
			name:     "IntSlice",
			input:    []int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:     "IntArray",
			input:    [3]int{1, 2, 3},
			expected: []interface{}{1, 2, 3},
		},
		{
			name:  "StringPointer",
			input: &str,
			error: "must be slice or array, got *string",
		},
		{
			name:     "SlicePointer",
			input:    &intSlice,
			expected: []interface{}{1, 2, 3},
		},
		{
			name: "JSONDoc",
			input: []interface{}{
				map[string]string{"foo": "bar"},
				[]byte(`{"foo":"bar"}`),
			},
			expected: []interface{}{
				map[string]string{"foo": "bar"},
				map[string]string{"foo": "bar"},
			},
		},
		{
			name: "BytesArrays",
			input: [][]byte{
				[]byte(`{"foo":"bar"}`),
				[]byte(`{"foo":"bar"}`),
			},
			expected: []interface{}{
				map[string]string{"foo": "bar"},
				map[string]string{"foo": "bar"},
			},
		},
		{
			name:  "InvalidJSON",
			input: []interface{}{[]byte(`invalid`)},
			error: "invalid character 'i' looking for beginning of value",
		},
		{
			name:  "BytesInvalidJSON",
			input: [][]byte{[]byte(`invalid`)},
			error: "invalid character 'i' looking for beginning of value",
		},
	}
	for _, test := range tests {
		func(test diTest) {
			t.Run(test.name, func(t *testing.T) {
				result, err := docsInterfaceSlice(test.input)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.error {
					t.Errorf("Unexpected error: %s", err)
				}
				if d := diff.AsJSON(test.expected, result); d != nil {
					t.Errorf("%s", d)
				}
			})
		}(test)
	}
}

func TestBulkDocsNotSlice(t *testing.T) {
	err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = r.(error)
			}
		}()
		db := &DB{}
		_, _ = db.BulkDocs(context.Background(), nil)
		return nil
	}()
	var msg string
	if err != nil {
		msg = err.Error()
	}
	expected := "must be slice or array, got <nil>"
	if msg != expected {
		t.Errorf("Unexpected error: %s", msg)
	}
}

type bdDB struct {
	driver.DB
	err     error
	options map[string]interface{}
}

var _ driver.DB = &bdDB{}

func (db *bdDB) BulkDocs(_ context.Context, docs []interface{}, options map[string]interface{}) (driver.BulkResults, error) {
	if db.options != nil {
		if !reflect.DeepEqual(db.options, options) {
			return nil, fmt.Errorf("Unexpected options. Got: %v, Expected: %v", options, db.options)
		}
	}
	return nil, db.err
}

type legacyDB struct {
	driver.DB
	err error
}

var _ driver.OldBulkDocer = &legacyDB{}

func (db *legacyDB) BulkDocs(_ context.Context, docs []interface{}) (driver.BulkResults, error) {
	return nil, db.err
}

type nonbdDB struct {
	driver.DB
}

var _ driver.DB = &nonbdDB{}

func (db *nonbdDB) Put(_ context.Context, _ string, _ interface{}) (string, error) { return "", nil }
func (db *nonbdDB) CreateDoc(_ context.Context, _ interface{}) (string, string, error) {
	return "", "", nil
}

func TestBulkDocs(t *testing.T) {
	type bdTest struct {
		name     string
		dbDriver driver.DB
		docs     interface{}
		options  Options
		err      string
	}
	tests := []bdTest{
		{
			name:     "no docs",
			dbDriver: &bdDB{},
			docs:     []int{},
		},
		{
			name:     "invalid JSON",
			dbDriver: &bdDB{},
			docs:     []interface{}{[]byte("invalid json")},
			err:      "invalid character 'i' looking for beginning of value",
		},
		{
			name:     "query fails",
			dbDriver: &bdDB{err: errors.New("bulkdocs failed")},
			docs:     []int{1, 2, 3},
			err:      "bulkdocs failed",
		},
		{
			name:     "emulated BulkDocs support",
			dbDriver: &nonbdDB{},
			docs: []interface{}{
				map[string]string{"_id": "foo"},
				123,
			},
		},
		{
			name:     "new_edits",
			dbDriver: &bdDB{options: map[string]interface{}{"new_edits": true}},
			docs: []interface{}{
				map[string]string{"_id": "foo"},
				123,
			},
			options: Options{"new_edits": true},
		},
		{
			name:     "legacy bulkDocer",
			dbDriver: &legacyDB{},
			docs: []interface{}{
				map[string]string{"_id": "foo"},
				123,
			},
		},
		{
			name:     "legacy failure",
			dbDriver: &legacyDB{err: errors.New("fail")},
			docs:     []interface{}{1, 2, 3},
			err:      "fail",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			db := &DB{driverDB: test.dbDriver}
			_, err := db.BulkDocs(context.Background(), test.docs, test.options)
			var msg string
			if err != nil {
				msg = err.Error()
			}
			if msg != test.err {
				t.Errorf("Unexpected error: %s", msg)
			}
		})
	}
}

func TestEmulatedBulkResults(t *testing.T) {
	results := []driver.BulkResult{
		{
			ID:    "chicken",
			Rev:   "foo",
			Error: nil,
		},
		{
			ID:    "duck",
			Rev:   "bar",
			Error: errors.New("fail"),
		},
		{
			ID:    "dog",
			Rev:   "baz",
			Error: nil,
		},
	}
	br := &emulatedBulkResults{results}
	result := &driver.BulkResult{}
	if err := br.Next(result); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if d := diff.Interface(&results[0], result); d != nil {
		t.Error(d)
	}
	if err := br.Next(result); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if d := diff.Interface(&results[1], result); d != nil {
		t.Error(d)
	}
	if err := br.Close(); err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
	if err := br.Next(result); err != io.EOF {
		t.Error("Expected EOF")
	}
}
