package kivik_test

import (
	"context"
	"net/http"

	"github.com/go-kivik/kivik/v3"
)

func ExampleStatusCode() {
	client, err := kivik.New("couch", "http://example.com:5984/")
	if err != nil {
		panic(err)
	}
	ctx := context.Background()
	row := client.DB(ctx, "foo").Get(ctx, "my_doc_id")
	switch kivik.StatusCode(row.Err) {
	case http.StatusNotFound:
		return
	case http.StatusUnauthorized:
		panic("Authentication required")
	case http.StatusForbidden:
		panic("You are not authorized")
	default:
		panic("Unexpected error: " + err.Error())
	}
}
