package test

import "github.com/flimzy/kivik/test/kt"

func init() {
	RegisterSuite(SuiteKivikFS, kt.SuiteConfig{
		"AllDBs.expected": []string{},
	})

}
