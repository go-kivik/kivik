package test

import "github.com/flimzy/kivik/test/kt"

func init() {
	RegisterSuite(SuiteCloudant, kt.SuiteConfig{
		"AllDBs.expected": []string{"_replicator", "_users"},
	})
}
