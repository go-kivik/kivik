package test

import (
	"testing"

	"github.com/flimzy/kivik"
)

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "Config", false, Config)
		RegisterTest(suite, "ConfigRW", true, ConfigRW)
	}
}

// Config tests the '/_config' endpoint
func Config(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		if suite == SuiteCloudant {
			testConfigRO(clients.Admin, kivik.StatusForbidden, t)
		} else {
			testConfigRO(clients.Admin, kivik.StatusNoError, t)
		}
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		testConfigRO(clients.NoAuth, kivik.StatusUnauthorized, t)
	})
}

func testConfigRO(client *kivik.Client, status int, t *testing.T) {
	c, err := client.Config()
	if err != nil {
		t.Errorf("Failed to get config object: %s", err)
		return
	}
	conf, err := c.GetAll()
	_ = IsError(err, status, t)
	if status == 0 {
		for _, section := range []string{"cors", "ssl", "httpd"} {
			if _, ok := conf[section]; !ok {
				t.Errorf("Config section '%s' missing", section)
			}
		}
	}
}

// ConfigRW tests the '/_config' endpoint with RW tests
func ConfigRW(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		if suite == SuiteCloudant {
			testConfigRW(clients.Admin, kivik.StatusForbidden, t)
		} else {
			testConfigRW(clients.Admin, kivik.StatusNoError, t)
		}
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		testConfigRW(clients.NoAuth, kivik.StatusUnauthorized, t)
	})
}

func testConfigRW(client *kivik.Client, status int, t *testing.T) {
	c, err := client.Config()
	if err != nil {
		t.Errorf("Failed to get config object: %s", err)
		return
	}

	// Now see if we can set a config option
	_ = IsError(c.Set("kivik", "kivik", "kivik"), status, t)
	_ = IsError(c.Delete("kivik", "kivik"), status, t)
	return
}
