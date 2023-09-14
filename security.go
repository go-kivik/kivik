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

package kivik

// Members represents the members of a database security document.
type Members struct {
	Names []string `json:"names,omitempty"`
	Roles []string `json:"roles,omitempty"`
}

// Security represents a database security document.
type Security struct {
	Admins  Members `json:"admins,omitempty"`
	Members Members `json:"members,omitempty"`

	// Database permissions for Cloudant users and/or API keys. This field is
	// only used or populated by IBM Cloudant. See the [Cloudant documentation]
	// for details.
	//
	// [Cloudant documentation]: https://cloud.ibm.com/apidocs/cloudant#getsecurity
	Cloudant map[string][]string `json:"cloudant,omitempty"`

	// Manage permissions using the `_users` database only. This field is only
	// used or populated by IBM Cloudant. See the [Cloudant documentation] for
	// details.
	//
	// [Cloudant documentation]: https://cloud.ibm.com/apidocs/cloudant#getsecurity
	CouchdbAuthOnly *bool `json:"couchdb_auth_only,omitempty"`
}
