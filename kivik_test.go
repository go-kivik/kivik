package kivik

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/testy"
)

func TestNew(t *testing.T) {
	registryMU.Lock()
	defer registryMU.Unlock()
	tests := []struct {
		name       string
		driver     driver.Driver
		driverName string
		dsn        string
		expected   *Client
		status     int
		err        string
	}{
		{
			name:       "Unregistered driver",
			driverName: "unregistered",
			dsn:        "unf",
			status:     StatusBadRequest,
			err:        `kivik: unknown driver "unregistered" (forgotten import?)`,
		},
		{
			name: "connection error",
			driver: &mockDriver{
				NewClientFunc: func(_ context.Context, _ string) (driver.Client, error) {
					return nil, errors.New("connection error")
				},
			},
			driverName: "foo",
			status:     StatusInternalServerError,
			err:        "connection error",
		},
		{
			name: "success",
			driver: &mockDriver{
				NewClientFunc: func(_ context.Context, dsn string) (driver.Client, error) {
					if dsn != "oink" {
						return nil, fmt.Errorf("Unexpected DSN: %s", dsn)
					}
					return &mockClient{id: "foo"}, nil
				},
			},
			driverName: "bar",
			dsn:        "oink",
			expected: &Client{
				dsn:          "oink",
				driverName:   "bar",
				driverClient: &mockClient{id: "foo"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			defer func() {
				drivers = make(map[string]driver.Driver)
			}()
			if test.driver != nil {
				Register(test.driverName, test.driver)
			}
			result, err := New(context.Background(), test.driverName, test.dsn)
			testy.StatusError(t, test.err, test.status, err)
			if d := diff.Interface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
