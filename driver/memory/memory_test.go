package memory

import (
	"context"
	"sort"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik/driver"
)

func TestNewClient(t *testing.T) {
	d := &memDriver{}
	_, err := d.NewClient(context.Background(), "foo")
	if err != nil {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestDBExists(t *testing.T) {
	d := &memDriver{}
	type deTest struct {
		Name     string
		DBName   string
		Setup    func(driver.Client)
		Expected bool
		Error    string
	}
	tests := []deTest{
		{
			Name:     "NoDBs",
			DBName:   "foo",
			Expected: false,
		},
		{
			Name:   "ExistingDB",
			DBName: "foo",
			Setup: func(c driver.Client) {
				if err := c.CreateDB(context.Background(), "foo", nil); err != nil {
					panic(err)
				}
			},
			Expected: true,
		},
		{
			Name:   "OtherDB",
			DBName: "foo",
			Setup: func(c driver.Client) {
				if err := c.CreateDB(context.Background(), "bar", nil); err != nil {
					panic(err)
				}
			},
			Expected: false,
		},
	}
	for _, test := range tests {
		func(test deTest) {
			t.Run(test.Name, func(t *testing.T) {
				c, err := d.NewClient(context.Background(), "foo")
				if err != nil {
					t.Fatal(err)
				}
				if test.Setup != nil {
					test.Setup(c)
				}
				result, err := c.DBExists(context.Background(), test.DBName, nil)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if test.Error != msg {
					t.Errorf("Unexpected error: %s", msg)
				}
				if result != test.Expected {
					t.Errorf("Expected: %t, Actual: %t", test.Expected, result)
				}
			})
		}(test)
	}
}

func TestCreateDB(t *testing.T) {
	d := &memDriver{}
	type cdTest struct {
		Name   string
		DBName string
		Error  string
		Setup  func(driver.Client)
	}
	tests := []cdTest{
		{
			Name:   "FirstDB",
			DBName: "foo",
		},
		{
			Name:   "Duplicate",
			DBName: "foo",
			Setup: func(c driver.Client) {
				if e := c.CreateDB(context.Background(), "foo", nil); e != nil {
					panic(e)
				}
			},
			Error: "database exists",
		},
	}
	for _, test := range tests {
		func(test cdTest) {
			t.Run(test.Name, func(t *testing.T) {
				c, err := d.NewClient(context.Background(), "foo")
				if err != nil {
					t.Fatal(err)
				}
				if test.Setup != nil {
					test.Setup(c)
				}
				var msg string
				if e := c.CreateDB(context.Background(), test.DBName, nil); e != nil {
					msg = e.Error()
				}
				if msg != test.Error {
					t.Errorf("Unexpected error: %s", msg)
				}
			})
		}(test)
	}
}

func TestAllDBs(t *testing.T) {
	d := &memDriver{}
	type adTest struct {
		Name     string
		Setup    func(driver.Client)
		Expected []string
		Error    string
	}
	tests := []adTest{
		{
			Name:     "NoDBs",
			Expected: []string{},
		},
		{
			Name: "2DBs",
			Setup: func(c driver.Client) {
				if err := c.CreateDB(context.Background(), "foo", nil); err != nil {
					panic(err)
				}
				if err := c.CreateDB(context.Background(), "bar", nil); err != nil {
					panic(err)
				}
			},
			Expected: []string{"foo", "bar"},
		},
	}
	for _, test := range tests {
		func(test adTest) {
			t.Run(test.Name, func(t *testing.T) {
				c, err := d.NewClient(context.Background(), "foo")
				if err != nil {
					t.Fatal(err)
				}
				if test.Setup != nil {
					test.Setup(c)
				}
				result, err := c.AllDBs(context.Background(), nil)
				var msg string
				if err != nil {
					msg = err.Error()
				}
				if msg != test.Error {
					t.Errorf("Unexpected error: %s", msg)
				}
				sort.Strings(test.Expected)
				sort.Strings(result)
				if d := diff.Interface(test.Expected, result); d != "" {
					t.Error(d)
				}
			})
		}(test)
	}
}
