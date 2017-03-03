// +build go1.7,!go1.8

package test

import (
	"os"
	"regexp"
	"testing"
)

func mainStart(clients *Clients, suites []string, rw bool) {
	m := testing.MainStart(regexp.MatchString, gatherTests(clients, suites, rw), nil, nil)
	os.Exit(m.Run())
}
