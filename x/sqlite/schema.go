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
	// revs
	`CREATE TABLE {{ .Revs }} (
		id TEXT NOT NULL,
		rev INTEGER NOT NULL,
		rev_id TEXT NOT NULL,
		parent_rev INTEGER,
		parent_rev_id TEXT,
		FOREIGN KEY (id, parent_rev, parent_rev_id) REFERENCES {{ .Revs }} (id, rev, rev_id) ON DELETE CASCADE,
		UNIQUE(id, rev, rev_id)
	)`,
	`CREATE INDEX idx_parent ON {{ .Revs }} (id, parent_rev, parent_rev_id)`,
	// the main db table
	`CREATE TABLE {{ .Docs }} (
		seq INTEGER PRIMARY KEY,
		id TEXT NOT NULL,
		rev INTEGER NOT NULL,
		rev_id TEXT NOT NULL,
		doc BLOB NOT NULL,
		deleted BOOLEAN NOT NULL DEFAULT FALSE,
		FOREIGN KEY (id, rev, rev_id) REFERENCES {{ .Revs }} (id, rev, rev_id) ON DELETE CASCADE,
		UNIQUE(id, rev, rev_id)
	)`,
	// attachments
	`CREATE TABLE {{ .Attachments }} (
		id TEXT NOT NULL,
		rev INTEGER NOT NULL,
		rev_id TEXT NOT NULL,
		filename TEXT NOT NULL,
		content_type TEXT NOT NULL,
		length INTEGER NOT NULL,
		digest TEXT NOT NULL,
		data BLOB NOT NULL,
		deleted_rev INTEGER,
		deleted_rev_id TEXT,
		FOREIGN KEY (id, rev, rev_id) REFERENCES {{ .Revs }} (id, rev, rev_id) ON DELETE CASCADE,
		UNIQUE(id, rev, rev_id, filename)
	)`,
	`CREATE VIEW {{ .Leaves }} AS
		SELECT
			doc.seq     AS seq,
			rev.id      AS id,
			rev.rev     AS rev,
			rev.rev_id  AS rev_id,
			doc.doc     AS doc,
			doc.deleted AS deleted
		FROM {{ .Revs }} AS rev
		LEFT JOIN {{ .Revs }} AS child ON rev.id = child.id AND rev.rev = child.parent_rev AND rev.rev_id = child.parent_rev_id
		JOIN {{ .Docs }} AS doc ON rev.id = doc.id AND rev.rev = doc.rev AND rev.rev_id = doc.rev_id
		WHERE child.id IS NULL
	`,
}
