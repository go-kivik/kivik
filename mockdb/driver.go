package mockdb

import (
	"errors"
	"fmt"
	"sync"
	"testing"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

var pool *mockDriver

func init() {
	pool = &mockDriver{
		clients: make(map[string]*Client),
	}
	kivik.Register("mock", pool)
}

type mockDriver struct {
	sync.Mutex
	counter int
	clients map[string]*Client
}

var _ driver.Driver = &mockDriver{}

func (d *mockDriver) NewClient(dsn string, _ driver.Options) (driver.Client, error) {
	d.Lock()
	defer d.Unlock()

	c, ok := d.clients[dsn]
	if !ok {
		return nil, errors.New("mockdb: no available connection found")
	}
	c.opened++
	return &driverClient{Client: c}, nil
}

// New creates a kivik client connection and a mock to manage expectations.
func New() (*kivik.Client, *Client, error) {
	pool.Lock()
	dsn := fmt.Sprintf("mockdb_%d", pool.counter)
	pool.counter++

	kmock := &Client{dsn: dsn, drv: pool, ordered: true}
	pool.clients[dsn] = kmock
	pool.Unlock()

	return kmock.open()
}

// NewT works exactly as New, except that any error will be passed to t.Fatal.
func NewT(t *testing.T) (*kivik.Client, *Client) {
	t.Helper()
	client, mock, err := New()
	if err != nil {
		t.Fatal(err)
	}
	return client, mock
}
