// +build ignore

package test

import (
	"net/http"
	"testing"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/errors"
)

func init() {
	for _, suite := range []string{SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "Membership", false, Membership)
	}
}

// Membership tests the /_membership endpoint
func Membership(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		testMembership(clients.Admin, suite, 0, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		if suite == SuiteCloudant {
			testMembership(clients.NoAuth, suite, http.StatusUnauthorized, t)
			return
		}
		testMembership(clients.NoAuth, suite, 0, t)
	})
}

func testMembership(client *kivik.Client, suite string, status int, t *testing.T) {
	t.Parallel()
	all, cluster := readMembership(client, status, t)
	if status > 0 {
		// Expected failure, so skip the rest
		return
	}
	if suite == SuiteCloudant {
		if len(all) < 2 {
			t.Errorf("Only got %d nodes, expected 2+\n", len(all))
		}
		if len(cluster) < 2 {
			t.Errorf("Only got %d cluster nodes, expected 2+\n", len(cluster))
		}
	}
	if len(all) < len(cluster) {
		t.Errorf("Cluster list (%d) shorter than full list (%d) (!!?!)", len(cluster), len(all))
	}
}

func readMembership(client *kivik.Client, status int, t *testing.T) (all, cluster []string) {
	var err error
	all, cluster, err = client.Membership()
	switch errors.StatusCode(err) {
	case status:
		return
	case 0:
		t.Errorf("Expected failure %d/%s", status, http.StatusText(status))
	default:
		t.Errorf("Failed to read membership: %s", err)
	}
	return
}
