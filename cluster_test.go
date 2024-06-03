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
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestClusterStatus(t *testing.T) {
	type tst struct {
		client   driver.Client
		closed   bool
		options  Option
		expected string
		status   int
		err      string
	}
	tests := testy.NewTable()
	tests.Add("driver doesn't implement Cluster interface", tst{
		client: &mock.Client{},
		status: http.StatusNotImplemented,
		err:    "kivik: driver does not support cluster operations",
	})
	tests.Add("client error", tst{
		client: &mock.Cluster{
			ClusterStatusFunc: func(context.Context, driver.Options) (string, error) {
				return "", errors.New("client error")
			},
		},
		status: http.StatusInternalServerError,
		err:    "client error",
	})
	tests.Add("success", tst{
		client: &mock.Cluster{
			ClusterStatusFunc: func(context.Context, driver.Options) (string, error) {
				return "cluster_finished", nil
			},
		},
		expected: "cluster_finished",
	})
	tests.Add("client closed", tst{
		client: &mock.Cluster{},
		closed: true,
		status: http.StatusServiceUnavailable,
		err:    "kivik: client closed",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		c := &Client{
			driverClient: test.client,
			closed:       test.closed,
		}
		result, err := c.ClusterStatus(context.Background(), test.options)
		if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
			t.Error(d)
		}
		if result != test.expected {
			t.Errorf("Unexpected status:\nExpected: %s\n  Actual: %s\n", test.expected, result)
		}
	})
}

func TestClusterSetup(t *testing.T) {
	type tst struct {
		client driver.Client
		closed bool
		action interface{}
		status int
		err    string
	}

	tests := testy.NewTable()
	tests.Add("driver doesn't implement Cluster interface", tst{
		client: &mock.Client{},
		status: http.StatusNotImplemented,
		err:    "kivik: driver does not support cluster operations",
	})
	tests.Add("client error", tst{
		client: &mock.Cluster{
			ClusterSetupFunc: func(context.Context, interface{}) error {
				return errors.New("client error")
			},
		},
		status: http.StatusInternalServerError,
		err:    "client error",
	})
	tests.Add("success", tst{
		client: &mock.Cluster{
			ClusterSetupFunc: func(context.Context, interface{}) error {
				return nil
			},
		},
	})
	tests.Add("client closed", tst{
		client: &mock.Cluster{},
		closed: true,
		status: http.StatusServiceUnavailable,
		err:    "kivik: client closed",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		c := &Client{
			driverClient: test.client,
			closed:       test.closed,
		}
		err := c.ClusterSetup(context.Background(), test.action)
		if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
			t.Error(d)
		}
	})
}

func TestMembership(t *testing.T) {
	type tt struct {
		client driver.Client
		closed bool
		want   *ClusterMembership
		status int
		err    string
	}

	tests := testy.NewTable()
	tests.Add("driver doesn't implement Cluster interface", tt{
		client: &mock.Client{},
		status: http.StatusNotImplemented,
		err:    "kivik: driver does not support cluster operations",
	})
	tests.Add("client error", tt{
		client: &mock.Cluster{
			MembershipFunc: func(context.Context) (*driver.ClusterMembership, error) {
				return nil, errors.New("client error")
			},
		},
		status: http.StatusInternalServerError,
		err:    "client error",
	})
	tests.Add("success", tt{
		client: &mock.Cluster{
			MembershipFunc: func(context.Context) (*driver.ClusterMembership, error) {
				return &driver.ClusterMembership{
					AllNodes:     []string{"one", "two", "three"},
					ClusterNodes: []string{"one", "two"},
				}, nil
			},
		},
		want: &ClusterMembership{
			AllNodes:     []string{"one", "two", "three"},
			ClusterNodes: []string{"one", "two"},
		},
	})
	tests.Add("client closed", tt{
		client: &mock.Cluster{},
		closed: true,
		status: http.StatusServiceUnavailable,
		err:    "kivik: client closed",
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		c := &Client{
			driverClient: tt.client,
			closed:       tt.closed,
		}
		got, err := c.Membership(context.Background())
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		if err != nil {
			return
		}
		if d := testy.DiffInterface(tt.want, got); d != nil {
			t.Error(d)
		}
	})
}
