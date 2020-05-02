package kivik_test

import (
	"context"
	"fmt"

	"github.com/go-kivik/kivik/v3"
)

var db = &kivik.DB{}

// Storing a document is done with Put or Create, which correspond to
// `PUT /{db}/{doc}` and `POST /{db}` respectively. In most cases, you should
// use Put.
func ExampleDB_store() {
	type Animal struct {
		ID       string `json:"_id"`
		Rev      string `json:"_rev,omitempty"`
		Feet     int    `json:"feet"`
		Greeting string `json:"greeting"`
	}

	cow := Animal{ID: "cow", Feet: 4, Greeting: "moo"}
	rev, err := db.Put(context.TODO(), "cow", cow)
	if err != nil {
		panic(err)
	}
	cow.Rev = rev
}

var cow struct {
	Rev      string
	Greeting string
}

// Updating a document is the same as storing one, except that the `_rev`
// parameter must match that stored on the server.
func ExampleDB_update() {
	cow.Rev = "1-6e609020e0371257432797b4319c5829" // Must be set
	cow.Greeting = "Moo!"
	newRev, err := db.Put(context.TODO(), "cow", cow)
	if err != nil {
		panic(err)
	}
	cow.Rev = newRev
}

// As with updating a document, deletion depends on the proper _rev parameter.
func ExampleDB_delete() {
	newRev, err := db.Delete(context.TODO(), "cow", "2-9c65296036141e575d32ba9c034dd3ee")
	if err != nil {
		panic(err)
	}
	fmt.Printf("The tombstone document has revision %s\n", newRev)
}

// When fetching a document, the document will be unmarshaled from JSON into
// your structure by the row.ScanDoc method.
func ExampleDB_fetch() {
	type Animal struct {
		ID       string `json:"_id"`
		Rev      string `json:"_rev,omitempty"`
		Feet     int    `json:"feet"`
		Greeting string `json:"greeting"`
	}

	var cow Animal
	err := db.Get(context.TODO(), "cow").ScanDoc(&cow)
	if err != nil {
		panic(err)
	}
	fmt.Printf("The cow says '%s'\n", cow.Greeting)
}

// Design documents are treated identically to normal documents by both CouchDB
// and Kivik. The only difference is the document ID.
//
// Store your document normally, formatted with your views (or other functions).
func ExampleDB_updateView() {
	_, err := db.Put(context.TODO(), "_design/foo", map[string]interface{}{
		"_id": "_design/foo",
		"views": map[string]interface{}{
			"foo_view": map[string]interface{}{
				"map": "function(doc) { emit(doc._id) }",
			},
		},
	})
	if err != nil {
		panic(err)
	}
}

func ExampleDB_query() {
	rows, err := db.Query(context.TODO(), "_design/foo", "_view/bar", kivik.Options{
		"startkey": `"foo"`,                           // Quotes are necessary so the
		"endkey":   `"foo` + kivik.EndKeySuffix + `"`, // key is a valid JSON object
	})
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		var doc interface{}
		if err := rows.ScanDoc(&doc); err != nil {
			panic(err)
		}
		/* do something with doc */
	}
	if rows.Err() != nil {
		panic(rows.Err())
	}
}

func ExampleDB_mapReduce() {
	opts := kivik.Options{
		"group": true,
	}
	rows, err := db.Query(context.TODO(), "_design/foo", "_view/bar", opts)
	if err != nil {
		panic(err)
	}
	for rows.Next() {
		/* ... */
	}
}
