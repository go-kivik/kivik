package memory

import (
	"testing"

	"golang.org/x/net/context"

	"github.com/flimzy/diff"
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
