package test

// SuiteConfig represents the configuration for a test suite.
type SuiteConfig map[string]interface{}

// Set sets a value.
func (c SuiteConfig) Set(key string, value interface{}) {
	c[key] = value
}

var suites = make(map[string]SuiteConfig)

// RegisterSuite registers a Suite as available for testing.
func RegisterSuite(suite string, conf SuiteConfig) {
	suites[suite] = conf
}
