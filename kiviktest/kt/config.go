// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package kt

import (
	"strings"
	"testing"

	kivik "github.com/go-kivik/kivik/v4"
)

// AllDocsIndex is the default index for _all_docs
var AllDocsIndex = kivik.Index{
	Name: "_all_docs", Type: "special",
	Definition: map[string]interface{}{"fields": []map[string]string{{"_id": "asc"}}},
}

// SuiteConfig represents the configuration for a test suite.
type SuiteConfig map[string]interface{}

func name(t *testing.T) string {
	name := t.Name()
	return name[strings.Index(name, "/")+1:]
}

// get looks for the requested key at the current level, and if not found it
// looks at the parent key.
func (c SuiteConfig) get(name, key string) interface{} {
	var k string
	if name == "" {
		k = key
	} else {
		k = name + "." + key
	}
	v, ok := c[k]
	if ok {
		return v
	}
	if name == "" {
		return nil
	}
	if !strings.Contains(name, "/") {
		return c.get("", key)
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
	if c.Bool(t, "skip") {
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

// String returns a string.
func (c SuiteConfig) String(t *testing.T, key string) string {
	v, _ := c.Interface(t, key).(string)
	return v
}
