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

	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func registerSuiteCouch33() {
	kiviktest.RegisterSuite(kiviktest.SuiteCouch33, mergeSuiteConfig(
		couch3xBase(),
		kt.SuiteConfig{
			"Version.version":                           `^3\.3\.`,
			"DBsStats/NoAuth.status":                    http.StatusUnauthorized,
			"DeleteAttachment/RW/Admin/NotFound.status": http.StatusNotFound,
			"Session/Post/BogusTypeJSON.status":         http.StatusUnsupportedMediaType,
			"Session/Post/BogusTypeForm.status":         http.StatusUnsupportedMediaType,
		},
	))
}
