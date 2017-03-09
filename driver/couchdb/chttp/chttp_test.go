package chttp

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func dsn() (string, bool) {
	for _, env := range []string{"KIVIK_TEST_DSN_COUCH16", "KIVIK_TEST_DSN_COUCH20", "KIVIK_TEST_DSN_CLOUDANT"} {
		dsn := os.Getenv(env)
		if dsn != "" {
			return dsn, true
		}
	}
	return "", false
}

func getClient(t *testing.T) *Client {
	dsn, ok := dsn()
	if !ok {
		t.Skip("No DSN set, skipping test.")
	}
	client, err := New(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to '%s': %s", dsn, err)
	}
	return client
}

func TestDo(t *testing.T) {
	client := getClient(t)
	res, err := client.Do("GET", "/", &Options{Accept: "application/json"})
	if err != nil {
		t.Errorf("Failed to make request GET /: %s", err)
	}
	body := &bytes.Buffer{}
	if _, err = body.ReadFrom(res.Body); err != nil {
		t.Errorf("Failed to read response body: %s", err)
	}
	if !strings.Contains(body.String(), `"couchdb"`) {
		t.Errorf("Body does not contain `\"couchdb\"` as expected: `%s`", body)
	}
	var i interface{}
	if err = json.Unmarshal(body.Bytes(), &i); err != nil {
		t.Errorf("Body is not valid JSON: %s", err)
	}
	if res.ContentType != "application/json" {
		t.Errorf("Unexpected content type: %s", res.ContentType)
	}
}

func TestJSONBody(t *testing.T) {
	client := getClient(t)
	bogusQuery := map[string]string{
		"foo": "bar",
		"bar": "baz",
	}
	_, err := client.Do("GET", "/_session", &Options{JSON: bogusQuery})
	if err != nil {
		t.Errorf("Failed to make request: %s", err)
	}
}
