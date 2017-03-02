package test

import (
	"regexp"

	"github.com/flimzy/kivik"
)

func init() {
	for _, suite := range AllSuites {
		RegisterTest(suite, "ServerInfo", false, ServerInfo)
	}
}

var versionREs = map[string]*regexp.Regexp{
	SuiteCouch:       regexp.MustCompile(`^1\.\d\.\d$`),
	SuiteCouch20:     regexp.MustCompile(`^2\.0\.0$`),
	SuiteKivikServer: regexp.MustCompile(`^1\.6\.1$`),
}

var vendorNames = map[string]string{
	SuiteCloudant:    "IBM Cloudant",
	SuiteCouch:       "The Apache Software Foundation",
	SuiteCouch20:     "The Apache Software Foundation",
	SuiteKivikMemory: "Kivik Memory Adaptor",
	SuitePouchLocal:  "PouchDB",
	SuiteKivikServer: "Kivik",
}

var vendorVersionREs = map[string]*regexp.Regexp{
	SuitePouchLocal:  regexp.MustCompile(`^\d\.\d\.\d$`),
	SuiteCouch:       regexp.MustCompile(`^1\.\d\.\d$`),
	SuiteCouch20:     regexp.MustCompile(`^2\.0\.0$`),
	SuiteKivikMemory: regexp.MustCompile(`^\d\.\d\.\d$`),
	SuiteCloudant:    regexp.MustCompile(`^\d\d\d\d$`),
	SuiteKivikServer: regexp.MustCompile(`^0\.0\.1$`),
}

// ServerInfo tests the '/' endpoint
func ServerInfo(client *kivik.Client, suite string, fail FailFunc) {
	info, err := client.ServerInfo()
	if err != nil {
		fail("%s", err)
		return
	}
	if re, ok := versionREs[suite]; ok {
		if !re.MatchString(info.Version()) {
			fail("Version %s does not match %s\n", info.Version(), re)
		}
	}
	if name, ok := vendorNames[suite]; ok {
		if name != info.Vendor() {
			fail("ServerInfo: Vendor Name %s does not match %s\n", info.Vendor(), name)
		}
	}
	if re, ok := vendorVersionREs[suite]; ok {
		if !re.MatchString(info.VendorVersion()) {
			fail("Vendor Version %s does not match %s\n", info.VendorVersion(), re)
		}
	}
}
