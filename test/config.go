package test

import (
	"net/http"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func init() {
	for _, suite := range []string{SuiteCouch16, SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "Config", false, Config)
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
	conf, err := client.GetAllConfig()
	switch errors.StatusCode(err) {
	case status:
		// expected
	case 0:
		t.Errorf("Expected failure %d/%s", status, http.StatusText(status))
	default:
		t.Errorf("Unexpected failure state.\nExpected: %d/%s\n  Actual: %s", status, http.StatusText(status), err)
	}
	if status == 0 {
		for _, section := range []string{"cors", "ssl", "httpd"} {
			if _, ok := conf[section]; !ok {
				t.Errorf("Config section '%s' missing", section)
			}
		}
	}
}
