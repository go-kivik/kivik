package test

import "github.com/flimzy/kivik/test/kt"

func init() {
	RegisterSuite(SuiteKivikMemory, kt.SuiteConfig{
		"AllDBs.expected": []string{},
	})

}
