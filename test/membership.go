package test

import "testing"

func init() {
	for _, suite := range []string{SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "Membership", false, Membership)
	}
}

// Membership tests the /_membership endpoint
func Membership(clients *Clients, suite string, t *testing.T) {
	client := clients.Admin
	all, cluster, err := client.Membership()
	if err != nil {
		t.Errorf("t.Errorfed to read membership: %s", err)
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
