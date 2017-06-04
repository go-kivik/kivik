package memory

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
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
				db, err := c.DB(context.Background(), test.DBName, nil)
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
				if d := diff.Interface(test.Expected, result); d != "" {
					t.Error(d)
				}
			})
		}(test)
	}
}

func setupDB(t *testing.T, s func(driver.DB)) driver.DB {
	c := setup(t, nil)
	if err := c.CreateDB(context.Background(), "foo", nil); err != nil {
		t.Fatal(err)
	}
	db, err := c.DB(context.Background(), "foo", nil)
	if err != nil {
		t.Fatal(err)
	}
	if s != nil {
		s(db)
	}
	return db
}

func TestPut(t *testing.T) {
	type putTest struct {
		Name   string
		DocID  string
		Doc    interface{}
		Setup  func(driver.DB)
		Status int
		Error  string
	}
	tests := []putTest{
		{
			Name:  "Success",
			DocID: "foo",
			Doc:   map[string]string{"_id": "foo", "foo": "bar"},
		},
		{
			Name:  "Conflict",
			DocID: "foo",
			Doc:   map[string]string{"_id": "foo", "_rev": "bar"},
			Setup: func(db driver.DB) {
				db.Put(context.Background(), "foo", map[string]string{"_id": "foo"})
			},
			Status: 409,
			Error:  "document update conflict",
		},
		{
			Name:  "InvalidJSON",
			DocID: "foo",
			Doc: func() interface{} {
				return map[string]interface{}{
					"channel": make(chan int),
				}
			}(),
			Status: 400,
			Error:  "invalid JSON",
		},
		{
			Name:   "InitialRev",
			DocID:  "foo",
			Doc:    map[string]string{"_id": "foo", "_rev": "bar"},
			Status: 409,
			Error:  "document update conflict",
		},
	}
	for _, test := range tests {
		func(test putTest) {
			t.Run(test.Name, func(t *testing.T) {
				db := setupDB(t, test.Setup)
				var msg string
				var status int
				if _, err := db.Put(context.Background(), test.DocID, test.Doc); err != nil {
					msg = err.Error()
					status = kivik.StatusCode(err)
				}
				if msg != test.Error {
					t.Errorf("Unexpected error: %s", msg)
				}
				if status != test.Status {
					t.Errorf("Unexpected status code: %d", status)
				}
			})
		}(test)
	}
}

// func TestCreateDoc(t *testing.T) {
// 	type cdTest struct {
// 		Name  string
// 		Doc   interface{}
// 		Error string
// 	}
// 	tests := []cdTest{
// 		{
// 			Name: "SimpleDoc",
// 			Doc: map[string]interface{}{
// 				"foo": "bar",
// 			},
// 		},
// 	}
// 	for _, test := range tests {
// 		func(test cdTest) {
// 			t.Run(test.Name, func(t *testing.T) {
// 				db := setupDB(t)
// 				_, _, err := db.CreateDoc(context.Background(), test.Doc)
// 				var msg string
// 				if err != nil {
// 					msg = err.Error()
// 				}
// 				if msg != test.Error {
// 					t.Errorf("Unexpected error: %s", msg)
// 				}
// 			})
// 		}(test)
// 	}
// }
