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
	"github.com/go-kivik/kivik/v4/internal/mock"
)

func TestClusterStatus(t *testing.T) {
	type tst struct {
		client   driver.Client
		options  Options
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
			ClusterStatusFunc: func(_ context.Context, _ map[string]interface{}) (string, error) {
				return "", errors.New("client error")
			},
		},
		status: http.StatusInternalServerError,
		err:    "client error",
	})
	tests.Add("success", tst{
		client: &mock.Cluster{
			ClusterStatusFunc: func(_ context.Context, _ map[string]interface{}) (string, error) {
				return "cluster_finished", nil
			},
		},
		expected: "cluster_finished",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		c := &Client{
			driverClient: test.client,
		}
		result, err := c.ClusterStatus(context.Background(), test.options)
		testy.StatusError(t, test.err, test.status, err)
		if result != test.expected {
			t.Errorf("Unexpected status:\nExpected: %s\n  Actual: %s\n", test.expected, result)
		}
	})
}

func TestClusterSetup(t *testing.T) {
	type tst struct {
		client driver.Client
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
			ClusterSetupFunc: func(_ context.Context, _ interface{}) error {
				return errors.New("client error")
			},
		},
		status: http.StatusInternalServerError,
		err:    "client error",
	})
	tests.Add("success", tst{
		client: &mock.Cluster{
			ClusterSetupFunc: func(_ context.Context, _ interface{}) error {
				return nil
			},
		},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		c := &Client{
			driverClient: test.client,
		}
		err := c.ClusterSetup(context.Background(), test.action)
		testy.StatusError(t, test.err, test.status, err)
	})
}

func TestMembership(t *testing.T) {
	type tt struct {
		client driver.Client
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
			MembershipFunc: func(_ context.Context) (*driver.ClusterMembership, error) {
				return nil, errors.New("client error")
			},
		},
		status: http.StatusInternalServerError,
		err:    "client error",
	})
	tests.Add("success", tt{
		client: &mock.Cluster{
			MembershipFunc: func(_ context.Context) (*driver.ClusterMembership, error) {
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

	tests.Run(t, func(t *testing.T, tt tt) {
		c := &Client{
			driverClient: tt.client,
		}
		got, err := c.Membership(context.Background())
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(tt.want, got); d != nil {
			t.Error(d)
		}
	})
}
