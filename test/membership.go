package test

import "github.com/flimzy/kivik"

func init() {
	for _, suite := range []string{SuiteCouch20, SuiteCloudant} {
		RegisterTest(suite, "Membership", false, Membership)
	}
}

// Membership tests the /_membership endpoint
func Membership(client *kivik.Client, suite string, fail FailFunc) {
	all, cluster, err := client.Membership()
	if err != nil {
		fail("Failed to read membership: %s", err)
	}
	if suite == SuiteCloudant {
		if len(all) < 2 {
			fail("Only got %d nodes, expected 2+\n", len(all))
		}
		if len(cluster) < 2 {
			fail("Only got %d cluster nodes, expected 2+\n", len(cluster))
		}
	}
	if len(all) < len(cluster) {
		fail("Cluster list (%d) shorter than full list (%d) (!!?!)", len(cluster), len(all))
	}
}
