// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package kivik

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestNew(t *testing.T) {
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
			status:     http.StatusBadRequest,
			err:        `kivik: unknown driver "unregistered" (forgotten import?)`,
		},
		{
			name: "connection error",
			driver: &mock.Driver{
				NewClientFunc: func(_ string, _ map[string]interface{}) (driver.Client, error) {
					return nil, errors.New("connection error")
				},
			},
			driverName: "foo",
			status:     http.StatusInternalServerError,
			err:        "connection error",
		},
		{
			name: "success",
			driver: &mock.Driver{
				NewClientFunc: func(dsn string, _ map[string]interface{}) (driver.Client, error) {
					if dsn != "oink" {
						return nil, fmt.Errorf("Unexpected DSN: %s", dsn)
					}
					return &mock.Client{ID: "foo"}, nil
				},
			},
			driverName: "bar",
			dsn:        "oink",
			expected: &Client{
				dsn:          "oink",
				driverName:   "bar",
				driverClient: &mock.Client{ID: "foo"},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if test.driver != nil {
				Register(test.driverName, test.driver)
			}
			result, err := New(test.driverName, test.dsn)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestClientGetters(t *testing.T) {
	driverName := "foo"
	dsn := "bar"
	c := &Client{
		driverName: driverName,
		dsn:        dsn,
	}

	t.Run("Driver", func(t *testing.T) {
		result := c.Driver()
		if result != driverName {
			t.Errorf("Unexpected result: %s", result)
		}
	})

	t.Run("DSN", func(t *testing.T) {
		result := c.DSN()
		if result != dsn {
			t.Errorf("Unexpected result: %s", result)
		}
	})
}

func TestVersion(t *testing.T) {
	tests := []struct {
		name     string
		client   *Client
		expected *Version
		status   int
		err      string
	}{
		{
			name: "db error",
			client: &Client{
				driverClient: &mock.Client{
					VersionFunc: func(_ context.Context) (*driver.Version, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			client: &Client{
				driverClient: &mock.Client{
					VersionFunc: func(_ context.Context) (*driver.Version, error) {
						return &driver.Version{Version: "foo"}, nil
					},
				},
			},
			expected: &Version{Version: "foo"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.Version(context.Background())
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestDB(t *testing.T) {
	type Test struct {
		name     string
		client   *Client
		dbName   string
		options  Options
		expected *DB
		status   int
		err      string
	}
	tests := []Test{
		{
			name: "db error",
			client: &Client{
				driverClient: &mock.Client{
					DBFunc: func(_ string, _ map[string]interface{}) (driver.DB, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		func() Test {
			client := &Client{
				driverClient: &mock.Client{
					DBFunc: func(dbName string, opts map[string]interface{}) (driver.DB, error) {
						expectedDBName := "foo"
						expectedOpts := map[string]interface{}{"foo": 123}
						if dbName != expectedDBName {
							return nil, fmt.Errorf("Unexpected dbname: %s", dbName)
						}
						if d := testy.DiffInterface(expectedOpts, opts); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return &mock.DB{ID: "abc"}, nil
					},
				},
			}
			return Test{
				name:    "success",
				client:  client,
				dbName:  "foo",
				options: map[string]interface{}{"foo": 123},
				expected: &DB{
					client:   client,
					name:     "foo",
					driverDB: &mock.DB{ID: "abc"},
				},
			}
		}(),
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.client.DB(test.dbName, test.options)
			err := result.Err()
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestAllDBs(t *testing.T) {
	tests := []struct {
		name     string
		client   *Client
		options  Options
		expected []string
		status   int
		err      string
	}{
		{
			name: "db error",
			client: &Client{
				driverClient: &mock.Client{
					AllDBsFunc: func(_ context.Context, _ map[string]interface{}) ([]string, error) {
						return nil, errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			client: &Client{
				driverClient: &mock.Client{
					AllDBsFunc: func(_ context.Context, options map[string]interface{}) ([]string, error) {
						expectedOptions := map[string]interface{}{"foo": 123}
						if d := testy.DiffInterface(expectedOptions, options); d != nil {
							return nil, fmt.Errorf("Unexpected options:\n%s", d)
						}
						return []string{"a", "b", "c"}, nil
					},
				},
			},
			options:  map[string]interface{}{"foo": 123},
			expected: []string{"a", "b", "c"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.AllDBs(context.Background(), test.options)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestDBExists(t *testing.T) {
	tests := []struct {
		name     string
		client   *Client
		dbName   string
		options  Options
		expected bool
		status   int
		err      string
	}{
		{
			name: "db error",
			client: &Client{
				driverClient: &mock.Client{
					DBExistsFunc: func(_ context.Context, _ string, _ map[string]interface{}) (bool, error) {
						return false, errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			client: &Client{
				driverClient: &mock.Client{
					DBExistsFunc: func(_ context.Context, dbName string, opts map[string]interface{}) (bool, error) {
						expectedDBName := "foo"
						expectedOpts := map[string]interface{}{"foo": 123}
						if dbName != expectedDBName {
							return false, fmt.Errorf("Unexpected db name: %s", dbName)
						}
						if d := testy.DiffInterface(expectedOpts, opts); d != nil {
							return false, fmt.Errorf("Unexpected opts:\n%s", d)
						}
						return true, nil
					},
				},
			},
			dbName:   "foo",
			options:  map[string]interface{}{"foo": 123},
			expected: true,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.DBExists(context.Background(), test.dbName, test.options)
			testy.StatusError(t, test.err, test.status, err)
			if test.expected != result {
				t.Errorf("Unexpected result: %v", result)
			}
		})
	}
}

func TestCreateDB(t *testing.T) {
	tests := []struct {
		name   string
		client *Client
		dbName string
		opts   Options
		status int
		err    string
	}{
		{
			name: "db error",
			client: &Client{
				driverClient: &mock.Client{
					CreateDBFunc: func(_ context.Context, _ string, _ map[string]interface{}) error {
						return errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			client: &Client{
				driverClient: &mock.Client{
					CreateDBFunc: func(_ context.Context, dbName string, opts map[string]interface{}) error {
						expectedDBName := "foo"
						expectedOpts := map[string]interface{}{"foo": 123}
						if dbName != expectedDBName {
							return fmt.Errorf("Unexpected dbname: %s", dbName)
						}
						if d := testy.DiffInterface(expectedOpts, opts); d != nil {
							return fmt.Errorf("Unexpected opts:\n%s", d)
						}
						return nil
					},
					DBFunc: func(dbName string, _ map[string]interface{}) (driver.DB, error) {
						return &mock.DB{ID: "abc"}, nil
					},
				},
			},
			dbName: "foo",
			opts:   map[string]interface{}{"foo": 123},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.client.CreateDB(context.Background(), test.dbName, test.opts)
			testy.StatusError(t, test.err, test.status, err)
		})
	}
}

func TestDestroyDB(t *testing.T) {
	tests := []struct {
		name   string
		client *Client
		dbName string
		opts   Options
		status int
		err    string
	}{
		{
			name: "db error",
			client: &Client{
				driverClient: &mock.Client{
					DestroyDBFunc: func(_ context.Context, _ string, _ map[string]interface{}) error {
						return errors.New("db error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "db error",
		},
		{
			name: "success",
			client: &Client{
				driverClient: &mock.Client{
					DestroyDBFunc: func(_ context.Context, dbName string, opts map[string]interface{}) error {
						expectedDBName := "foo"
						expectedOpts := map[string]interface{}{"foo": 123}
						if dbName != expectedDBName {
							return fmt.Errorf("Unexpected dbname: %s", dbName)
						}
						if d := testy.DiffInterface(expectedOpts, opts); d != nil {
							return fmt.Errorf("Unexpected opts:\n%s", d)
						}
						return nil
					},
				},
			},
			dbName: "foo",
			opts:   map[string]interface{}{"foo": 123},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.client.DestroyDB(context.Background(), test.dbName, test.opts)
			testy.StatusError(t, test.err, test.status, err)
		})
	}
}

func TestAuthenticate(t *testing.T) {
	tests := []struct {
		name   string
		client *Client
		auth   interface{}
		status int
		err    string
	}{
		{
			name: "non-authenticator",
			client: &Client{
				driverClient: &mock.Client{},
			},
			status: http.StatusNotImplemented,
			err:    "kivik: driver does not support authentication",
		},
		{
			name: "auth error",
			client: &Client{
				driverClient: &mock.Authenticator{
					AuthenticateFunc: func(_ context.Context, _ interface{}) error {
						return errors.New("auth error")
					},
				},
			},
			status: http.StatusInternalServerError,
			err:    "auth error",
		},
		{
			name: "success",
			client: &Client{
				driverClient: &mock.Authenticator{
					AuthenticateFunc: func(_ context.Context, a interface{}) error {
						expected := int(3)
						if d := testy.DiffInterface(expected, a); d != nil {
							return fmt.Errorf("Unexpected authenticator:\n%s", d)
						}
						return nil
					},
				},
			},
			auth: int(3),
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := test.client.Authenticate(context.Background(), test.auth)
			testy.StatusError(t, test.err, test.status, err)
		})
	}
}

func TestDBsStats(t *testing.T) {
	tests := []struct {
		name     string
		client   *Client
		dbnames  []string
		expected []*DBStats
		err      string
		status   int
	}{
		{
			name: "fallback to old driver",
			client: &Client{
				driverClient: &mock.Client{
					DBFunc: func(name string, _ map[string]interface{}) (driver.DB, error) {
						switch name {
						case "foo":
							return &mock.DB{
								StatsFunc: func(_ context.Context) (*driver.DBStats, error) {
									return &driver.DBStats{Name: "foo", DiskSize: 123}, nil
								},
							}, nil
						case "bar":
							return &mock.DB{
								StatsFunc: func(_ context.Context) (*driver.DBStats, error) {
									return &driver.DBStats{Name: "bar", DiskSize: 321}, nil
								},
							}, nil
						default:
							return nil, errors.New("not found")
						}
					},
				},
			},
			dbnames: []string{"foo", "bar"},
			expected: []*DBStats{
				{Name: "foo", DiskSize: 123},
				{Name: "bar", DiskSize: 321},
			},
		},
		{
			name: "fallback due to old server",
			client: &Client{
				driverClient: &mock.Client{
					DBFunc: func(name string, _ map[string]interface{}) (driver.DB, error) {
						switch name {
						case "foo":
							return &mock.DB{
								StatsFunc: func(_ context.Context) (*driver.DBStats, error) {
									return &driver.DBStats{Name: "foo", DiskSize: 123}, nil
								},
							}, nil
						case "bar":
							return &mock.DB{
								StatsFunc: func(_ context.Context) (*driver.DBStats, error) {
									return &driver.DBStats{Name: "bar", DiskSize: 321}, nil
								},
							}, nil
						default:
							return nil, errors.New("not found")
						}
					},
				},
			},
			dbnames: []string{"foo", "bar"},
			expected: []*DBStats{
				{Name: "foo", DiskSize: 123},
				{Name: "bar", DiskSize: 321},
			},
		},
		{
			name: "native success",
			client: &Client{
				driverClient: &mock.DBsStatser{
					DBsStatsFunc: func(_ context.Context, names []string) ([]*driver.DBStats, error) {
						return []*driver.DBStats{
							{Name: "foo", DiskSize: 123},
							{Name: "bar", DiskSize: 321},
						}, nil
					},
				},
			},
			dbnames: []string{"foo", "bar"},
			expected: []*DBStats{
				{Name: "foo", DiskSize: 123},
				{Name: "bar", DiskSize: 321},
			},
		},
		{
			name: "native error",
			client: &Client{
				driverClient: &mock.DBsStatser{
					DBsStatsFunc: func(_ context.Context, names []string) ([]*driver.DBStats, error) {
						return nil, errors.New("native failure")
					},
				},
			},
			dbnames: []string{"foo", "bar"},
			err:     "native failure",
			status:  500,
		},
		{
			name: "fallback error",
			client: &Client{
				driverClient: &mock.Client{
					DBFunc: func(_ string, _ map[string]interface{}) (driver.DB, error) {
						return &mock.DB{
							StatsFunc: func(_ context.Context) (*driver.DBStats, error) {
								return nil, errors.New("fallback failure")
							},
						}, nil
					},
				},
			},
			dbnames: []string{"foo", "bar"},
			err:     "fallback failure",
			status:  500,
		},
		{
			name: "fallback db connect error",
			client: &Client{
				driverClient: &mock.Client{
					DBFunc: func(_ string, _ map[string]interface{}) (driver.DB, error) {
						return nil, errors.New("db conn failure")
					},
				},
			},
			dbnames: []string{"foo", "bar"},
			err:     "db conn failure",
			status:  500,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			stats, err := test.client.DBsStats(context.Background(), test.dbnames)
			testy.StatusError(t, test.err, test.status, err)
			if d := testy.DiffInterface(test.expected, stats); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestPing(t *testing.T) {
	type pingTest struct {
		name     string
		client   *Client
		expected bool
		err      string
	}
	tests := []pingTest{
		{
			name: "non-pinger",
			client: &Client{
				driverClient: &mock.Client{
					VersionFunc: func(_ context.Context) (*driver.Version, error) {
						return &driver.Version{}, nil
					},
				},
			},
			expected: true,
		},
		{
			name: "pinger",
			client: &Client{
				driverClient: &mock.Pinger{
					PingFunc: func(_ context.Context) (bool, error) {
						return true, nil
					},
				},
			},
			expected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.client.Ping(context.Background())
			testy.Error(t, test.err, err)
			if result != test.expected {
				t.Errorf("Unexpected result: %t", result)
			}
		})
	}
}

func TestMergeOptions(t *testing.T) {
	type tst struct {
		options  []Options
		expected Options
	}
	tests := testy.NewTable()
	tests.Add("No options", tst{})
	tests.Add("One set", tst{
		options: []Options{
			{"foo": 123},
		},
		expected: Options{"foo": 123},
	})
	tests.Add("merged", tst{
		options: []Options{
			{"foo": 123},
			{"bar": 321},
		},
		expected: Options{
			"foo": 123,
			"bar": 321,
		},
	})
	tests.Add("overwrite", tst{
		options: []Options{
			{"foo": 123, "bar": 321},
			{"foo": 111},
		},
		expected: Options{
			"foo": 111,
			"bar": 321,
		},
	})
	tests.Add("nil option", tst{
		options: []Options{nil},
	})
	tests.Add("different types", tst{
		options: []Options{
			{"foo": 123},
			{"foo": "bar"},
		},
		expected: Options{"foo": "bar"},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result := mergeOptions(test.options...)
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestClientClose(t *testing.T) {
	type tst struct {
		client *Client
		err    string
	}
	tests := testy.NewTable()
	tests.Add("non-closer", tst{
		client: &Client{driverClient: &mock.Client{}},
	})
	tests.Add("error", tst{
		client: &Client{driverClient: &mock.ClientCloser{
			CloseFunc: func(_ context.Context) error {
				return errors.New("close err")
			},
		}},
		err: "close err",
	})
	tests.Add("success", tst{
		client: &Client{driverClient: &mock.ClientCloser{
			CloseFunc: func(_ context.Context) error {
				return nil
			},
		}},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		err := test.client.Close(context.Background())
		testy.Error(t, test.err, err)
	})
}
