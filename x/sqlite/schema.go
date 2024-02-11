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

package sqlite

var schema = []string{
	`CREATE TABLE %q (
		seq INTEGER PRIMARY KEY,
		id TEXT NOT NULL,
		rev INTEGER NOT NULL DEFAULT 1,
		rev_id TEXT NOT NULL,
		doc BLOB NOT NULL,
		deleted BOOLEAN NOT NULL DEFAULT FALSE,
		UNIQUE(id, rev, rev_id)
	)`,
}
