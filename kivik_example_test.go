package kivik_test

import (
	"context"
	"fmt"

	"github.com/go-kivik/kivik/v3"
	"github.com/go-kivik/kivik/v3/driver"
	"github.com/go-kivik/kivik/v3/internal/mock"
)

func init() {
	kivik.Register("couch", &mock.Driver{
		NewClientFunc: func(_ string) (driver.Client, error) {
			return &mock.Client{
				DBFunc: func(_ context.Context, _ string, _ map[string]interface{}) (driver.DB, error) {
					return nil, nil
				},
			}, nil
		},
	})
}

// New is used to create a client handle. `driver` specifies the name of the
// registered database driver and `dataSourceName` specifies the
// database-specific connection information, such as a URL.
func ExampleNew() {
	client, err := kivik.New("couch", "http://example.com:5984/")
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to", client.DSN())
	// Output: Connected to http://example.com:5984/
}

// With a client handle in hand, you can create a database handle with the DB()
// method to interact with a specific database.
func Example_connecting() {
	client, err := kivik.New("couch", "http://example.com:5984/")
	if err != nil {
		panic(err)
	}
	db := client.DB(context.TODO(), "_users")
	fmt.Println("Database handle for " + db.Name())
	// Output: Database handle for _users
}
