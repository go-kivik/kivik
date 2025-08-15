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

// Package test manages the integration tests for the PostgreSQL driver.
package test

import (
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

// RegisterPGSuites registers the PostgreSQL related integration test suites.
func RegisterPGSuites() {
	registerSuitePG()
}

func registerSuitePG() {
	kiviktest.RegisterSuite(kiviktest.SuitePG, kt.SuiteConfig{})
}
