package kt

import (
	"strings"
	"testing"
)

// SuiteConfig represents the configuration for a test suite.
type SuiteConfig map[string]interface{}

func name(t *testing.T) string {
	name := tName(t)
	return name[strings.Index(name, "/")+1:]
}

// get looks for the requested key at the current level, and if not found it
// looks at the parent key.
func (c SuiteConfig) get(name, key string) interface{} {
	v, ok := c[name+"."+key]
	if ok {
		return v
	}
	if !strings.Contains(name, "/") {
		return nil
	}
	// Try the parent
	return c.get(name[0:strings.LastIndex(name, "/")], key)
}

// Interface returns the configuration value as an interface{}.
func (c SuiteConfig) Interface(t *testing.T, key string) interface{} {
	return c.get(name(t), key)
}

// Bool returns the boolean value of the key.
func (c SuiteConfig) Bool(t *testing.T, key string) bool {
	b, _ := c.Interface(t, key).(bool)
	return b
}

// Skip will skip the currently running test if configuration dictates.
func (c SuiteConfig) Skip(t *testing.T) {
	if c.Bool(t, "Skip") {
		t.Skip("Test skipped by suite configuration")
	}
}

// StringSlice returns a string slice.
func (c SuiteConfig) StringSlice(t *testing.T, key string) []string {
	v, _ := c.Interface(t, key).([]string)
	return v
}

// Int returns an int.
func (c SuiteConfig) Int(t *testing.T, key string) int {
	v, _ := c.Interface(t, key).(int)
	return v
}
