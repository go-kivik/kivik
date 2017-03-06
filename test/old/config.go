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
	t.Run("GetAll", func(t *testing.T) {
		t.Parallel()
		conf, err := c.GetAll()
		_ = IsError(err, status, t)
		if status == 0 {
			for _, section := range []string{"cors", "ssl", "httpd"} {
				if _, ok := conf[section]; !ok {
					t.Errorf("Config section '%s' missing", section)
				}
			}
		}
	})
	t.Run("GetSection", func(t *testing.T) {
		t.Parallel()
		sec, err := c.GetSection("couch_httpd_auth")
		_ = IsError(err, status, t)
		if status == 0 {
			for _, key := range []string{"allow_persistent_cookies", "iterations"} {
				if _, ok := sec[key]; !ok {
					t.Errorf("Config key 'couch_httpd_auth.%s' missing", key)
				}
			}
		}
	})
	t.Run("Get", func(t *testing.T) {
		t.Parallel()
		value, err := c.Get("couchdb", "file_compression")
		_ = IsError(err, status, t)
		if status == 0 {
			if value != "snappy" {
				t.Errorf("Value couchdb.file_compression = '%s', expected 'snappy'", value)
			}
		}
	})
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
	if status == 0 {
		value, err := c.Get("kivik", "kivik")
		_ = IsError(err, 0, t)
		if value != "kivik" {
			t.Errorf("Retrieved kivik.kivik = '%s' after setting to 'kivik'", value)
		}
	}
	_ = IsError(c.Delete("kivik", "kivik"), status, t)
	return
}
