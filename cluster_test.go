package kivik

import (
	"context"
	"errors"
	"testing"

	"github.com/flimzy/testy"

	"github.com/go-kivik/kivik/driver"
	"github.com/go-kivik/kivik/mock"
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
		status: StatusNotImplemented,
		err:    "kivik: driver does not support cluster operations",
	})
	tests.Add("client error", tst{
		client: &mock.Cluster{
			ClusterStatusFunc: func(_ context.Context, _ map[string]interface{}) (string, error) {
				return "", errors.New("client error")
			},
		},
		status: StatusInternalServerError,
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
		status: StatusNotImplemented,
		err:    "kivik: driver does not support cluster operations",
	})
	tests.Add("client error", tst{
		client: &mock.Cluster{
			ClusterSetupFunc: func(_ context.Context, _ interface{}) error {
				return errors.New("client error")
			},
		},
		status: StatusInternalServerError,
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
