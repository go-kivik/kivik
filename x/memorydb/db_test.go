package memorydb

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

func TestStats(t *testing.T) {
	type statTest struct {
		Name     string
		DBName   string
		Setup    func(driver.Client)
		Expected *driver.DBStats
		Error    string
	}
	tests := []statTest{
		{
			Name:   "NoDBs",
			DBName: "foo",
			Setup: func(c driver.Client) {
				if e := c.CreateDB(context.Background(), "foo", nil); e != nil {
					panic(e)
				}
			},
			Expected: &driver.DBStats{Name: "foo"},
		},
	}
	for _, test := range tests {
		func(test statTest) {
			t.Run(test.Name, func(t *testing.T) {
				c := setup(t, test.Setup)
				db, err := c.DB(test.DBName, nil)
				if err != nil {
					t.Fatal(err)
				}
				result, err := db.Stats(context.Background())
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.Error {
					t.Errorf("Unexpected error: %s", msg)
				}
				if err != nil {
					return
				}
				if d := testy.DiffInterface(test.Expected, result); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

func setupDB(t *testing.T) *db {
	c := setup(t, nil)
	if err := c.CreateDB(context.Background(), "foo", nil); err != nil {
		t.Fatal(err)
	}
	d, err := c.DB("foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	return d.(*db)
}

func TestPut(t *testing.T) {
	type putTest struct {
		Name     string
		DocID    string
		Doc      interface{}
		Setup    func() *db
		Expected interface{}
		Status   int
		Error    string
	}
	tests := []putTest{
		{
			Name:   "LeadingUnderscoreInID",
			DocID:  "_badid",
			Doc:    map[string]string{"_id": "_badid", "foo": "bar"},
			Status: 400,
			Error:  "only reserved document ids may start with underscore",
		},
		{
			Name:     "MismatchedIDs",
			DocID:    "foo",
			Doc:      map[string]string{"_id": "bar"},
			Expected: map[string]string{"_id": "foo", "_rev": "1-xxx"},
		},
		{
			Name:     "Success",
			DocID:    "foo",
			Doc:      map[string]string{"_id": "foo", "foo": "bar"},
			Expected: map[string]string{"_id": "foo", "foo": "bar", "_rev": "1-xxx"},
		},
		{
			Name:  "Conflict",
			DocID: "foo",
			Doc:   map[string]string{"_id": "foo", "_rev": "bar"},
			Setup: func() *db {
				db := setupDB(t)
				if _, err := db.Put(context.Background(), "foo", map[string]string{"_id": "foo"}, nil); err != nil {
					t.Fatal(err)
				}
				return db
			},
			Status: 409,
			Error:  "document update conflict",
		},
		{
			Name:  "Unmarshalable",
			DocID: "foo",
			Doc: func() interface{} {
				return map[string]interface{}{
					"channel": make(chan int),
				}
			}(),
			Status: 400,
			Error:  "json: unsupported type: chan int",
		},
		{
			Name:   "InitialRev",
			DocID:  "foo",
			Doc:    map[string]string{"_id": "foo", "_rev": "bar"},
			Status: 409,
			Error:  "document update conflict",
		},
		func() putTest {
			dbv := setupDB(t)
			rev, err := dbv.Put(context.Background(), "foo", map[string]string{"_id": "foo", "foo": "bar"}, nil)
			if err != nil {
				panic(err)
			}
			return putTest{
				Name:     "Update",
				DocID:    "foo",
				Setup:    func() *db { return dbv },
				Doc:      map[string]string{"_id": "foo", "_rev": rev},
				Expected: map[string]string{"_id": "foo", "_rev": "2-xxx"},
			}
		}(),
		{
			Name:     "DesignDoc",
			DocID:    "_design/foo",
			Doc:      map[string]string{"foo": "bar"},
			Expected: map[string]string{"_id": "_design/foo", "foo": "bar", "_rev": "1-xxx"},
		},
		{
			Name:     "LocalDoc",
			DocID:    "_local/foo",
			Doc:      map[string]string{"foo": "bar"},
			Expected: map[string]string{"_id": "_local/foo", "foo": "bar", "_rev": "1-0"},
		},
		{
			Name:     "RecreateDeleted",
			DocID:    "foo",
			Doc:      map[string]string{"foo": "bar"},
			Expected: map[string]string{"_id": "foo", "foo": "bar", "_rev": "3-xxx"},
			Setup: func() *db {
				db := setupDB(t)
				rev, err := db.Put(context.Background(), "foo", map[string]string{"_id": "foo"}, nil)
				if err != nil {
					t.Fatal(err)
				}
				if _, e := db.Delete(context.Background(), "foo", kivik.Rev(rev)); e != nil {
					t.Fatal(e)
				}
				return db
			},
		},
		{
			Name:     "LocalDoc",
			DocID:    "_local/foo",
			Doc:      map[string]string{"foo": "baz"},
			Expected: map[string]string{"_id": "_local/foo", "foo": "baz", "_rev": "1-0"},
			Setup: func() *db {
				db := setupDB(t)
				_, err := db.Put(context.Background(), "_local/foo", map[string]string{"foo": "bar"}, nil)
				if err != nil {
					t.Fatal(err)
				}
				return db
			},
		},
		{
			Name:  "WithAttachments",
			DocID: "duck",
			Doc: map[string]interface{}{
				"_id":   "duck",
				"value": "quack",
				"_attachments": []map[string]interface{}{
					{"foo.css": map[string]string{
						"content_type": "text/css",
						"data":         "LyogYW4gZW1wdHkgQ1NTIGZpbGUgKi8=",
					}},
				},
			},
			Expected: map[string]string{
				"_id":   "duck",
				"_rev":  "1-xxx",
				"value": "quack",
			},
		},
		{
			Name: "Deleted DB",
			Setup: func() *db {
				c := setup(t, nil)
				if err := c.CreateDB(context.Background(), "deleted0", nil); err != nil {
					t.Fatal(err)
				}
				dbv, err := c.DB("deleted0", nil)
				if err != nil {
					t.Fatal(err)
				}
				if e := c.DestroyDB(context.Background(), "deleted0", nil); e != nil {
					t.Fatal(e)
				}
				return dbv.(*db)
			},
			Status: http.StatusPreconditionFailed,
			Error:  "database does not exist",
		},
	}
	for _, test := range tests {
		func(test putTest) {
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()
				var db *db
				if test.Setup != nil {
					db = test.Setup()
				} else {
					db = setupDB(t)
				}
				_, err := db.Put(context.Background(), test.DocID, test.Doc, nil)
				testy.StatusError(t, test.Error, test.Status, err)
				doc, err := db.Get(context.Background(), test.DocID, kivik.Params(nil))
				if err != nil {
					t.Fatal(err)
				}
				var result map[string]interface{}
				if e := json.NewDecoder(doc.Body).Decode(&result); e != nil {
					t.Fatal(e)
				}
				if !strings.HasPrefix(test.DocID, "_local/") {
					if rev, ok := result["_rev"].(string); ok {
						parts := strings.SplitN(rev, "-", 2)
						result["_rev"] = parts[0] + "-xxx"
					}
				}
				if d := testy.DiffAsJSON(test.Expected, result); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestGet(t *testing.T) {
	type getTest struct {
		Name     string
		ID       string
		options  kivik.Option
		DB       *db
		Status   int
		Error    string
		doc      *driver.Document
		Expected interface{}
	}
	tests := []getTest{
		{
			Name:   "NotFound",
			ID:     "foo",
			Status: 404,
			Error:  "missing",
		},
		func() getTest {
			db := setupDB(t)
			rev, err := db.Put(context.Background(), "foo", map[string]string{"_id": "foo", "foo": "bar"}, nil)
			if err != nil {
				panic(err)
			}
			return getTest{
				Name: "ExistingDoc",
				ID:   "foo",
				DB:   db,
				doc: &driver.Document{
					Rev: rev,
				},
				Expected: map[string]string{"_id": "foo", "foo": "bar"},
			}
		}(),
		func() getTest {
			db := setupDB(t)
			rev, err := db.Put(context.Background(), "foo", map[string]string{"_id": "foo", "foo": "Bar"}, nil)
			if err != nil {
				panic(err)
			}
			return getTest{
				Name:    "SpecificRev",
				ID:      "foo",
				DB:      db,
				options: kivik.Rev(rev),
				doc: &driver.Document{
					Rev: rev,
				},
				Expected: map[string]string{"_id": "foo", "foo": "Bar"},
			}
		}(),
		func() getTest {
			db := setupDB(t)
			rev, err := db.Put(context.Background(), "foo", map[string]string{"_id": "foo", "foo": "Bar"}, nil)
			if err != nil {
				panic(err)
			}
			_, err = db.Put(context.Background(), "foo", map[string]string{"_id": "foo", "foo": "baz", "_rev": rev}, nil)
			if err != nil {
				panic(err)
			}
			return getTest{
				Name:    "OldRev",
				ID:      "foo",
				DB:      db,
				options: kivik.Rev(rev),
				doc: &driver.Document{
					Rev: rev,
				},
				Expected: map[string]string{"_id": "foo", "foo": "Bar"},
			}
		}(),
		{
			Name:    "MissingRev",
			ID:      "foo",
			options: kivik.Rev("1-4c6114c65e295552ab1019e2b046b10e"),
			DB: func() *db {
				db := setupDB(t)
				_, err := db.Put(context.Background(), "foo", map[string]string{"_id": "foo", "foo": "Bar"}, nil)
				if err != nil {
					panic(err)
				}
				return db
			}(),
			Status: 404,
			Error:  "missing",
		},
		func() getTest {
			db := setupDB(t)
			rev, err := db.Put(context.Background(), "foo", map[string]string{"_id": "foo"}, nil)
			if err != nil {
				panic(err)
			}
			if _, e := db.Delete(context.Background(), "foo", kivik.Rev(rev)); e != nil {
				panic(e)
			}
			return getTest{
				Name:   "DeletedDoc",
				ID:     "foo",
				DB:     db,
				Status: 404,
				Error:  "missing",
			}
		}(),
		{
			Name: "Deleted DB",
			DB: func() *db {
				c := setup(t, nil)
				if err := c.CreateDB(context.Background(), "deleted0", nil); err != nil {
					t.Fatal(err)
				}
				dbv, err := c.DB("deleted0", nil)
				if err != nil {
					t.Fatal(err)
				}
				if e := c.DestroyDB(context.Background(), "deleted0", nil); e != nil {
					t.Fatal(e)
				}
				return dbv.(*db)
			}(),
			Error:  "database does not exist",
			Status: http.StatusPreconditionFailed,
		},
	}
	for _, test := range tests {
		func(test getTest) {
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()
				db := test.DB
				if db == nil {
					db = setupDB(t)
				}
				opts := test.options
				if opts == nil {
					opts = kivik.Params(nil)
				}
				doc, err := db.Get(context.Background(), test.ID, opts)
				testy.StatusError(t, test.Error, test.Status, err)
				var result map[string]interface{}
				if err := json.NewDecoder(doc.Body).Decode(&result); err != nil {
					t.Fatal(err)
				}
				doc.Body = nil // Determinism
				if d := testy.DiffInterface(test.doc, doc); d != nil {
					t.Errorf("Unexpected doc:\n%s", d)
				}
				if result != nil {
					delete(result, "_rev")
				}
				if d := testy.DiffAsJSON(test.Expected, result); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestDeleteDoc(t *testing.T) {
	type delTest struct {
		Name   string
		ID     string
		Rev    string
		DB     *db
		Status int
		Error  string
	}
	tests := []delTest{
		{
			Name:   "NonExistingDoc",
			ID:     "foo",
			Rev:    "1-4c6114c65e295552ab1019e2b046b10e",
			Status: 404,
			Error:  "missing",
		},
		func() delTest {
			db := setupDB(t)
			rev, err := db.Put(context.Background(), "foo", map[string]string{"_id": "foo"}, nil)
			if err != nil {
				panic(err)
			}
			return delTest{
				Name: "Success",
				ID:   "foo",
				DB:   db,
				Rev:  rev,
			}
		}(),
		{
			Name:   "InvalidRevFormat",
			ID:     "foo",
			Rev:    "invalid rev format",
			Status: 400,
			Error:  "invalid rev format",
		},
		{
			Name: "LocalNoRev",
			ID:   "_local/foo",
			Rev:  "",
			DB: func() *db {
				db := setupDB(t)
				if _, err := db.Put(context.Background(), "_local/foo", map[string]string{"foo": "bar"}, nil); err != nil {
					panic(err)
				}
				return db
			}(),
		},
		{
			Name: "LocalWithRev",
			ID:   "_local/foo",
			Rev:  "0-1",
			DB: func() *db {
				db := setupDB(t)
				if _, err := db.Put(context.Background(), "_local/foo", map[string]string{"foo": "bar"}, nil); err != nil {
					panic(err)
				}
				return db
			}(),
		},
		{
			Name: "DB deleted",
			ID:   "foo",
			DB: func() *db {
				c := setup(t, nil)
				if err := c.CreateDB(context.Background(), "deleted0", nil); err != nil {
					t.Fatal(err)
				}
				dbv, err := c.DB("deleted0", nil)
				if err != nil {
					t.Fatal(err)
				}
				if e := c.DestroyDB(context.Background(), "deleted0", nil); e != nil {
					t.Fatal(e)
				}
				return dbv.(*db)
			}(),
			Status: http.StatusPreconditionFailed,
			Error:  "database does not exist",
		},
	}
	for _, test := range tests {
		func(test delTest) {
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()
				db := test.DB
				if db == nil {
					db = setupDB(t)
				}
				rev, err := db.Delete(context.Background(), test.ID, kivik.Rev(test.Rev))
				var msg string
				var status int
				if err != nil {
					msg = err.Error()
					status = kivik.HTTPStatus(err)
				}
				if msg != test.Error {
					t.Errorf("Unexpected error: %s", msg)
				}
				if status != test.Status {
					t.Errorf("Unexpected status: %d", status)
				}
				if err != nil {
					return
				}
				row, err := db.Get(context.Background(), test.ID, kivik.Rev(rev))
				if err != nil {
					t.Fatal(err)
				}
				var doc interface{}
				if e := json.NewDecoder(row.Body).Decode(&doc); e != nil {
					t.Fatal(e)
				}
				expected := map[string]interface{}{
					"_id":      test.ID,
					"_rev":     rev,
					"_deleted": true,
				}
				if d := testy.DiffAsJSON(expected, doc); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestCreateDoc(t *testing.T) {
	type cdTest struct {
		Name     string
		DB       *db
		Doc      interface{}
		Expected map[string]interface{}
		Error    string
	}
	tests := []cdTest{
		{
			Name: "SimpleDoc",
			Doc: map[string]interface{}{
				"foo": "bar",
			},
			Expected: map[string]interface{}{
				"_rev": "1-xxx",
				"foo":  "bar",
			},
		},
		{
			Name: "Deleted DB",
			DB: func() *db {
				c := setup(t, nil)
				if err := c.CreateDB(context.Background(), "deleted0", nil); err != nil {
					t.Fatal(err)
				}
				dbv, err := c.DB("deleted0", nil)
				if err != nil {
					t.Fatal(err)
				}
				if e := c.DestroyDB(context.Background(), "deleted0", nil); e != nil {
					t.Fatal(e)
				}
				return dbv.(*db)
			}(),
			Doc: map[string]interface{}{
				"foo": "bar",
			},
			Error: "database does not exist",
		},
	}
	for _, test := range tests {
		func(test cdTest) {
			t.Run(test.Name, func(t *testing.T) {
				db := test.DB
				if db == nil {
					db = setupDB(t)
				}
				docID, _, err := db.CreateDoc(context.Background(), test.Doc, nil)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.Error {
					t.Errorf("Unexpected error: %s", msg)
				}
				if err != nil {
					return
				}
				row, err := db.Get(context.Background(), docID, kivik.Params(nil))
				if err != nil {
					t.Fatal(err)
				}
				var result map[string]interface{}
				if e := json.NewDecoder(row.Body).Decode(&result); e != nil {
					t.Fatal(e)
				}
				if result["_id"].(string) != docID {
					t.Errorf("Unexpected id. %s != %s", result["_id"].(string), docID)
				}
				delete(result, "_id")
				if rev, ok := result["_rev"].(string); ok {
					parts := strings.SplitN(rev, "-", 2)
					result["_rev"] = parts[0] + "-xxx"
				}
				if d := testy.DiffInterface(test.Expected, result); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}
