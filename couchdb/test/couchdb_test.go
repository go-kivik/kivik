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

//go:build !js
// +build !js

package test

import (
	"testing"

	_ "github.com/go-kivik/couchdb/v4"
	"github.com/go-kivik/kivik/v4/kiviktest"
)

func init() {
	RegisterCouchDBSuites()
}

func TestCouch16(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch16, "KIVIK_TEST_DSN_COUCH16")
}

func TestCouch17(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch17, "KIVIK_TEST_DSN_COUCH17")
}

func TestCouch20(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch20, "KIVIK_TEST_DSN_COUCH20")
}

func TestCouch21(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch21, "KIVIK_TEST_DSN_COUCH21")
}

func TestCouch22(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch22, "KIVIK_TEST_DSN_COUCH22")
}

func TestCouch23(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch23, "KIVIK_TEST_DSN_COUCH23")
}

func TestCouch30(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch30, "KIVIK_TEST_DSN_COUCH30")
}

func TestCouch31(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch31, "KIVIK_TEST_DSN_COUCH31")
}

func TestCouch32(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch32, "KIVIK_TEST_DSN_COUCH32")
}

func TestCouch33(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCouch33, "KIVIK_TEST_DSN_COUCH33")
}

func TestCloudant(t *testing.T) {
	kiviktest.DoTest(t, kiviktest.SuiteCloudant, "KIVIK_TEST_DSN_CLOUDANT")
}
