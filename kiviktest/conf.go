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

// Package kiviktest provides integration tests for kivik.
//
// This package is considered internal to the Kivik project, and is not subject
// to semver rules. It may change at any time, without any notice. Use at your
// own risk. You have been warned!
package kiviktest

import "github.com/go-kivik/kivik/v4/kiviktest/kt"

var suites = make(map[string]kt.SuiteConfig)

// RegisterSuite registers a Suite as available for testing.
func RegisterSuite(suite string, conf kt.SuiteConfig) {
	if _, ok := suites[suite]; ok {
		panic(suite + " already registered")
	}
	suites[suite] = conf
}
