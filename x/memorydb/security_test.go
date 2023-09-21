package memorydb

import (
	"context"
	"testing"

	"github.com/go-kivik/kivik/v4/driver"
	"gitlab.com/flimzy/testy"
)

func TestGetSecurity(t *testing.T) {
	type secTest struct {
		Name     string
		DB       *db
		Error    string
		Expected interface{}
	}
	tests := []secTest{
		{
			Name:  "DBNotFound",
			Error: "database does not exist",
			DB: func() *db {
				c := setup(t, nil)
				if err := c.CreateDB(context.Background(), "foo", nil); err != nil {
					t.Fatal(err)
				}
				dbv, err := c.DB("foo", nil)
				if err != nil {
					t.Fatal(err)
				}
				if e := c.DestroyDB(context.Background(), "foo", nil); e != nil {
					t.Fatal(e)
				}
				return dbv.(*db)
			}(),
		},
		{
			Name: "EmptySecurity",
			DB: func() *db {
				c := setup(t, nil)
				if err := c.CreateDB(context.Background(), "foo", nil); err != nil {
					t.Fatal(err)
				}
				dbv, err := c.DB("foo", nil)
				if err != nil {
					t.Fatal(err)
				}
				return dbv.(*db)
			}(),
			Expected: &driver.Security{},
		},
		{
			Name: "AdminsAndMembers",
			DB: func() *db {
				c := setup(t, nil)
				if err := c.CreateDB(context.Background(), "foo", nil); err != nil {
					t.Fatal(err)
				}

				db := &db{
					dbName: "foo",
					client: c.(*client),
					db: &database{
						security: &driver.Security{
							Admins: driver.Members{
								Names: []string{"foo", "bar", "baz"},
								Roles: []string{"morons"},
							},
							Members: driver.Members{
								Names: []string{"bob"},
								Roles: []string{"boring"},
							},
						},
					},
				}
				return db
			}(),
			Expected: &driver.Security{
				Admins: driver.Members{
					Names: []string{"foo", "bar", "baz"},
					Roles: []string{"morons"},
				},
				Members: driver.Members{
					Names: []string{"bob"},
					Roles: []string{"boring"},
				},
			},
		},
	}
	for _, test := range tests {
		func(test secTest) {
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()
				db := test.DB
				if db == nil {
					db = setupDB(t)
				}
				sec, err := db.Security(context.Background())
				testy.Error(t, test.Error, err)
				if d := testy.DiffAsJSON(test.Expected, sec); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}

func TestSetSecurity(t *testing.T) {
	type setTest struct {
		Name     string
		Security *driver.Security
		Error    string
		Expected *driver.Security
		DB       *db
	}
	tests := []setTest{
		{
			Name:  "DBNotFound",
			Error: "missing",
			DB: func() *db {
				c := setup(t, nil)
				if err := c.CreateDB(context.Background(), "foo", nil); err != nil {
					t.Fatal(err)
				}
				dbv, err := c.DB("foo", nil)
				if err != nil {
					t.Fatal(err)
				}
				if e := c.DestroyDB(context.Background(), "foo", nil); e != nil {
					t.Fatal(e)
				}
				return dbv.(*db)
			}(),
		},
		{
			Name: "Valid",
			Security: &driver.Security{
				Admins: driver.Members{
					Names: []string{"foo", "bar", "baz"},
					Roles: []string{"morons"},
				},
				Members: driver.Members{
					Names: []string{"bob"},
					Roles: []string{"boring"},
				},
			},
			Expected: &driver.Security{
				Admins: driver.Members{
					Names: []string{"foo", "bar", "baz"},
					Roles: []string{"morons"},
				},
				Members: driver.Members{
					Names: []string{"bob"},
					Roles: []string{"boring"},
				},
			},
		},
	}
	for _, test := range tests {
		func(test setTest) {
			t.Run(test.Name, func(t *testing.T) {
				t.Parallel()
				db := test.DB
				if db == nil {
					db = setupDB(t)
				}
				err := db.SetSecurity(context.Background(), test.Security)
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
				sec, err := db.Security(context.Background())
				if err != nil {
					t.Fatal(err)
				}
				if d := testy.DiffAsJSON(test.Expected, sec); d != nil {
					t.Error(d)
				}
			})
		}(test)
	}
}
