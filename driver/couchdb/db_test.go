package couchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/url"
	"strings"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
)

func TestAllDocs(t *testing.T) {
	client := getClient(t)
	db, err := client.DB(context.Background(), "_users", kivik.Options{"force_commit": true})
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	rows, err := db.AllDocs(context.Background(), map[string]interface{}{
		"include_docs": true,
	})
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}

	for {
		err := rows.Next(&driver.Row{})
		if err != nil {
			if err != io.EOF {
				t.Fatalf("Iteration failed: %s", err)
			}
			break
		}
	}
}

func TestDBInfo(t *testing.T) {
	client := getClient(t)
	db, err := client.DB(context.Background(), "_users", kivik.Options{"force_commit": true})
	if err != nil {
		t.Fatalf("Failed to connect to db: %s", err)
	}
	info, err := db.Stats(context.Background())
	if err != nil {
		t.Fatalf("Failed: %s", err)
	}
	if info.Name != "_users" {
		t.Errorf("Unexpected name %s", info.Name)
	}
}

func TestOptionsToParams(t *testing.T) {
	type otpTest struct {
		Name     string
		Input    map[string]interface{}
		Expected url.Values
		Error    string
	}
	tests := []otpTest{
		{
			Name:     "String",
			Input:    map[string]interface{}{"foo": "bar"},
			Expected: map[string][]string{"foo": []string{"bar"}},
		},
		{
			Name:     "StringSlice",
			Input:    map[string]interface{}{"foo": []string{"bar", "baz"}},
			Expected: map[string][]string{"foo": []string{"bar", "baz"}},
		},
		{
			Name:     "Bool",
			Input:    map[string]interface{}{"foo": true},
			Expected: map[string][]string{"foo": []string{"true"}},
		},
		{
			Name:     "Int",
			Input:    map[string]interface{}{"foo": 123},
			Expected: map[string][]string{"foo": []string{"123"}},
		},
		{
			Name:  "Error",
			Input: map[string]interface{}{"foo": []byte("foo")},
			Error: "cannot convert type []uint8 to []string",
		},
	}
	for _, test := range tests {
		func(test otpTest) {
			t.Run(test.Name, func(t *testing.T) {
				params, err := optionsToParams(test.Input)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.Error {
					t.Errorf("Error\n\tExpected: %s\n\t  Actual: %s\n", test.Error, msg)
				}
				if d := diff.Interface(test.Expected, params); d != "" {
					t.Errorf("Params not as expected:\n%s\n", d)
				}
			})
		}(test)
	}
}

func TestJSONify(t *testing.T) {
	type jsonifyTest struct {
		Name     string
		Input    interface{}
		Expected string
	}
	tests := []jsonifyTest{
		{
			Name:     "Null",
			Expected: "null",
		},
		{
			Name:     "String",
			Input:    `{"foo":"bar"}`,
			Expected: `{"foo":"bar"}`,
		},
		{
			Name:     "ByteSlice",
			Input:    []byte(`{"foo":"bar"}`),
			Expected: `{"foo":"bar"}`,
		},
		{
			Name:     "RawMessage",
			Input:    json.RawMessage(`{"foo":"bar"}`),
			Expected: `{"foo":"bar"}`,
		},
		{
			Name:     "Interface",
			Input:    map[string]string{"foo": "bar"},
			Expected: `{"foo":"bar"}`,
		},
	}
	for _, test := range tests {
		func(test jsonifyTest) {
			t.Run(test.Name, func(t *testing.T) {
				r, err := jsonify(test.Input)
				if err != nil {
					t.Fatalf("jsonify failed: %s", err)
				}
				buf := &bytes.Buffer{}
				buf.ReadFrom(r)
				result := strings.TrimSpace(buf.String())
				if result != test.Expected {
					t.Errorf("Expected: `%s`\n  Actual: `%s`", test.Expected, result)
				}
			})
		}(test)
	}
}
