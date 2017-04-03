package chttp

import (
	"bytes"
	"context"
	"encoding/json"
	"mime"
	"os"
	"strings"
	"testing"
)

func dsn(t *testing.T) string {
	for _, env := range []string{"KIVIK_TEST_DSN_COUCH16", "KIVIK_TEST_DSN_COUCH20", "KIVIK_TEST_DSN_CLOUDANT"} {
		dsn := os.Getenv(env)
		if dsn != "" {
			return dsn
		}
	}
	t.Skip("DSN not set")
	return ""
}

func getClient(t *testing.T) *Client {
	dsn := dsn(t)
	client, err := New(dsn)
	if err != nil {
		t.Fatalf("Failed to connect to '%s': %s", dsn, err)
	}
	return client
}

func TestDo(t *testing.T) {
	client := getClient(t)
	res, err := client.DoReq(context.Background(), "GET", "/", &Options{Accept: "application/json"})
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
	ct, _, err := mime.ParseMediaType(res.Header.Get("Content-Type"))
	if err != nil {
		t.Errorf("Failed to parse content type: %s", err)
	}
	if ct != "application/json" {
		t.Errorf("Unexpected content type: %s", ct)
	}
}

func TestJSONBody(t *testing.T) {
	client := getClient(t)
	bogusQuery := map[string]string{
		"foo": "bar",
		"bar": "baz",
	}
	buf := &bytes.Buffer{}
	if err := json.NewEncoder(buf).Encode(bogusQuery); err != nil {
		t.Fatalf("JSON encoding failed: %s", err)
	}
	_, err := client.DoReq(context.Background(), "POST", "/_session", &Options{Body: buf})
	if err != nil {
		t.Errorf("Failed to make request: %s", err)
	}
}
