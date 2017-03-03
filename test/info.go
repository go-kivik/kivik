package test

import (
	"regexp"
	"testing"

	"github.com/flimzy/kivik"
)

func init() {
	for _, suite := range AllSuites {
		RegisterTest(suite, "ServerInfo", false, ServerInfo)
	}
}

var versionREs = map[string]*regexp.Regexp{
	SuiteCouch16:     regexp.MustCompile(`^1\.\d\.\d$`),
	SuiteCouch20:     regexp.MustCompile(`^2\.0\.0$`),
	SuiteKivikServer: regexp.MustCompile(`^1\.6\.1$`),
}

var vendorNames = map[string]string{
	SuiteCloudant:    "IBM Cloudant",
	SuiteCouch16:     "The Apache Software Foundation",
	SuiteCouch20:     "The Apache Software Foundation",
	SuiteKivikMemory: "Kivik Memory Adaptor",
	SuitePouchLocal:  "PouchDB",
	SuiteKivikServer: "Kivik",
}

var vendorVersionREs = map[string]*regexp.Regexp{
	SuitePouchLocal:  regexp.MustCompile(`^\d\.\d\.\d$`),
	SuiteCouch16:     regexp.MustCompile(`^1\.\d\.\d$`),
	SuiteCouch20:     regexp.MustCompile(`^2\.0\.0$`),
	SuiteKivikMemory: regexp.MustCompile(`^\d\.\d\.\d$`),
	SuiteCloudant:    regexp.MustCompile(`^\d\d\d\d$`),
	SuiteKivikServer: regexp.MustCompile(`^0\.0\.1$`),
}

// ServerInfo tests the '/' endpoint
func ServerInfo(clients *Clients, suite string, t *testing.T) {
	t.Run("Admin", func(t *testing.T) {
		testServerInfo(clients.Admin, suite, t)
	})
	if clients.NoAuth == nil {
		return
	}
	t.Run("NoAuth", func(t *testing.T) {
		testServerInfo(clients.NoAuth, suite, t)
	})
}

func testServerInfo(client *kivik.Client, suite string, t *testing.T) {
	t.Parallel()
	info, err := client.ServerInfo()
	if err != nil {
		t.Errorf("%s", err)
		return
	}
	if re, ok := versionREs[suite]; ok {
		if !re.MatchString(info.Version()) {
			t.Errorf("Version %s does not match %s\n", info.Version(), re)
		}
	}
	if name, ok := vendorNames[suite]; ok {
		if name != info.Vendor() {
			t.Errorf("ServerInfo: Vendor Name %s does not match %s\n", info.Vendor(), name)
		}
	}
	if re, ok := vendorVersionREs[suite]; ok {
		if !re.MatchString(info.VendorVersion()) {
			t.Errorf("Vendor Version %s does not match %s\n", info.VendorVersion(), re)
		}
	}
}
