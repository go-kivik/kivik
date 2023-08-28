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

package test

import (
	"net/http"
)

const maxConnsPerHost = 90 // 100

// RegisterCouchDBSuites registers the CouchDB related integration test suites.
func RegisterCouchDBSuites() {
	registerSuiteCouch20()
	registerSuiteCouch21()
	registerSuiteCouch22()
	registerSuiteCouch23()
	registerSuiteCouch30()
	registerSuiteCouch31()
	registerSuiteCouch32()
	registerSuiteCouch33()
	registerSuiteCloudant()
}

func httpClient() *http.Client {
	client := &http.Client{
		Transport: http.DefaultTransport,
	}
	client.Transport.(*http.Transport).MaxConnsPerHost = maxConnsPerHost
	return client
}
