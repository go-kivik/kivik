package kivik

import (
	"context"
	"net/http"
	"testing"

	"github.com/flimzy/diff"
	"github.com/flimzy/testy"
	"github.com/go-kivik/kivik/driver"
	"github.com/go-kivik/kivik/mock"
	"github.com/pkg/errors"
)

func TestConfig(t *testing.T) {
	type tst struct {
		client   *Client
		node     string
		expected Config
		status   int
		err      string
	}
	tests := testy.NewTable()
	tests.Add("non-configer", tst{
		client: &Client{driverClient: &mock.Client{}},
		status: http.StatusNotImplemented,
		err:    "kivik: driver does not support Config interface",
	})
	tests.Add("error", tst{
		client: &Client{driverClient: &mock.Configer{
			ConfigFunc: func(_ context.Context, _ string) (driver.Config, error) {
				return nil, errors.New("conf error")
			},
		}},
		status: http.StatusInternalServerError,
		err:    "conf error",
	})
	tests.Add("success", tst{
		client: &Client{driverClient: &mock.Configer{
			ConfigFunc: func(_ context.Context, node string) (driver.Config, error) {
				if node != "foo" {
					return nil, errors.Errorf("Unexpected node: %s", node)
				}
				return driver.Config{
					"foo": driver.ConfigSection{"asd": "rew"},
				}, nil
			},
		}},
		node: "foo",
		expected: Config{
			"foo": ConfigSection{"asd": "rew"},
		},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.Config(context.Background(), test.node)
		testy.StatusError(t, test.err, test.status, err)
		if d := diff.Interface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestConfigSection(t *testing.T) {
	type tst struct {
		client        *Client
		node, section string
		expected      ConfigSection
		status        int
		err           string
	}
	tests := testy.NewTable()
	tests.Add("non-configer", tst{
		client: &Client{driverClient: &mock.Client{}},
		status: http.StatusNotImplemented,
		err:    "kivik: driver does not support Config interface",
	})
	tests.Add("error", tst{
		client: &Client{driverClient: &mock.Configer{
			ConfigSectionFunc: func(_ context.Context, _, _ string) (driver.ConfigSection, error) {
				return nil, errors.New("conf error")
			},
		}},
		status: http.StatusInternalServerError,
		err:    "conf error",
	})
	tests.Add("success", tst{
		client: &Client{driverClient: &mock.Configer{
			ConfigSectionFunc: func(_ context.Context, node, section string) (driver.ConfigSection, error) {
				if node != "foo" {
					return nil, errors.Errorf("Unexpected node: %s", node)
				}
				if section != "foo" {
					return nil, errors.Errorf("Unexpected section: %s", section)
				}
				return driver.ConfigSection{"lkj": "ghj"}, nil
			},
		}},
		node:     "foo",
		section:  "foo",
		expected: ConfigSection{"lkj": "ghj"},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.ConfigSection(context.Background(), test.node, test.section)
		testy.StatusError(t, test.err, test.status, err)
		if d := diff.Interface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestConfigValue(t *testing.T) {
	type tst struct {
		client             *Client
		node, section, key string
		expected           string
		status             int
		err                string
	}
	tests := testy.NewTable()
	tests.Add("non-configer", tst{
		client: &Client{driverClient: &mock.Client{}},
		status: http.StatusNotImplemented,
		err:    "kivik: driver does not support Config interface",
	})
	tests.Add("error", tst{
		client: &Client{driverClient: &mock.Configer{
			ConfigValueFunc: func(_ context.Context, _, _, _ string) (string, error) {
				return "", errors.New("conf error")
			},
		}},
		status: http.StatusInternalServerError,
		err:    "conf error",
	})
	tests.Add("success", tst{
		client: &Client{driverClient: &mock.Configer{
			ConfigValueFunc: func(_ context.Context, node, section, key string) (string, error) {
				if node != "foo" {
					return "", errors.Errorf("Unexpected node: %s", node)
				}
				if section != "foo" {
					return "", errors.Errorf("Unexpected section: %s", section)
				}
				if key != "asd" {
					return "", errors.Errorf("Unexpected key: %s", key)
				}
				return "jkl", nil
			},
		}},
		node:     "foo",
		section:  "foo",
		key:      "asd",
		expected: "jkl",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.ConfigValue(context.Background(), test.node, test.section, test.key)
		testy.StatusError(t, test.err, test.status, err)
		if d := diff.Interface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestSetConfigValue(t *testing.T) {
	type tst struct {
		client                    *Client
		node, section, key, value string
		expected                  string
		status                    int
		err                       string
	}
	tests := testy.NewTable()
	tests.Add("non-configer", tst{
		client: &Client{driverClient: &mock.Client{}},
		status: http.StatusNotImplemented,
		err:    "kivik: driver does not support Config interface",
	})
	tests.Add("error", tst{
		client: &Client{driverClient: &mock.Configer{
			SetConfigValueFunc: func(_ context.Context, _, _, _, _ string) (string, error) {
				return "", errors.New("conf error")
			},
		}},
		status: http.StatusInternalServerError,
		err:    "conf error",
	})
	tests.Add("success", tst{
		client: &Client{driverClient: &mock.Configer{
			SetConfigValueFunc: func(_ context.Context, node, section, key, value string) (string, error) {
				if node != "foo" {
					return "", errors.Errorf("Unexpected node: %s", node)
				}
				if section != "foo" {
					return "", errors.Errorf("Unexpected section: %s", section)
				}
				if key != "vbn" {
					return "", errors.Errorf("Unexpected key: %s", key)
				}
				if value != "baz" {
					return "", errors.Errorf("Unexpected value: %s", value)
				}
				return "old", nil
			},
		}},
		node:     "foo",
		section:  "foo",
		key:      "vbn",
		value:    "baz",
		expected: "old",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.SetConfigValue(context.Background(), test.node, test.section, test.key, test.value)
		testy.StatusError(t, test.err, test.status, err)
		if d := diff.Interface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

func TestDeleteConfigKey(t *testing.T) {
	type tst struct {
		client             *Client
		node, section, key string
		expected           string
		status             int
		err                string
	}
	tests := testy.NewTable()
	tests.Add("non-configer", tst{
		client: &Client{driverClient: &mock.Client{}},
		status: http.StatusNotImplemented,
		err:    "kivik: driver does not support Config interface",
	})
	tests.Add("error", tst{
		client: &Client{driverClient: &mock.Configer{
			DeleteConfigKeyFunc: func(_ context.Context, _, _, _ string) (string, error) {
				return "", errors.New("conf error")
			},
		}},
		status: http.StatusInternalServerError,
		err:    "conf error",
	})
	tests.Add("success", tst{
		client: &Client{driverClient: &mock.Configer{
			DeleteConfigKeyFunc: func(_ context.Context, node, section, key string) (string, error) {
				if node != "foo" {
					return "", errors.Errorf("Unexpected node: %s", node)
				}
				if section != "foo" {
					return "", errors.Errorf("Unexpected section: %s", section)
				}
				if key != "baz" {
					return "", errors.Errorf("Unexpected key: %s", key)
				}
				return "old", nil
			},
		}},
		node:     "foo",
		section:  "foo",
		key:      "baz",
		expected: "old",
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := test.client.DeleteConfigKey(context.Background(), test.node, test.section, test.key)
		testy.StatusError(t, test.err, test.status, err)
		if d := diff.Interface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}
