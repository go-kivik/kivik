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
		key TEXT GENERATED ALWAYS AS (json_quote(id)) VIRTUAL COLLATE COUCHDB_UCI,
		parent_rev INTEGER,
		parent_rev_id TEXT,
		FOREIGN KEY (id, parent_rev, parent_rev_id) REFERENCES {{ .Revs }} (id, rev, rev_id) ON DELETE CASCADE,
		UNIQUE(id, rev, rev_id)
	)`,
	`CREATE INDEX default_key ON {{ .Revs }} (key)`,
	`CREATE INDEX idx_parent ON {{ .Revs }} (id, parent_rev, parent_rev_id)`,
	// the main db table
	`CREATE TABLE {{ .Docs }} (
		seq INTEGER PRIMARY KEY,
		id TEXT NOT NULL,
		rev INTEGER NOT NULL,
		rev_id TEXT NOT NULL,
		doc BLOB NOT NULL,
		md5sum BLOB NOT NULL,
		deleted BOOLEAN NOT NULL DEFAULT FALSE,
		FOREIGN KEY (id, rev, rev_id) REFERENCES {{ .Revs }} (id, rev, rev_id) ON DELETE CASCADE,
		UNIQUE(id, rev, rev_id)
	)`,
	// attachments
	`CREATE TABLE {{ .Attachments }} (
		pk INTEGER PRIMARY KEY,
		filename TEXT NOT NULL,
		content_type TEXT NOT NULL,
		length INTEGER NOT NULL,
		digest BLOB NOT NULL,
		data BLOB NOT NULL,
		rev_pos INTEGER NOT NULL
	)`,
	`CREATE TABLE {{ .AttachmentsBridge }} (
		pk INTEGER,
		id TEXT NOT NULL,
		rev INTEGER NOT NULL,
		rev_id TEXT NOT NULL,
		FOREIGN KEY (pk) REFERENCES {{ .Attachments }} (pk),
		FOREIGN KEY (id, rev, rev_id) REFERENCES {{ .Docs }} (id, rev, rev_id) ON DELETE CASCADE,
		UNIQUE (id, rev, rev_id, pk)
	)`,
	/*
		The .Design table is used to store design documents. The schema is as follows:
		- id: The document ID.
		- rev: The revision number.
		- rev_id: The revision ID.  id, rev, and rev_id together form the primary key, which is also a foreign key to the .Docs table.
		- language: The language of the design document. Defaults to 'javascript'. Duplicated for each function, for convenience when doing function lookups.
		- func_type: The function type. One of 'map', 'reduce', 'update', 'filter', or 'validate', for use as a view map or reduce function respectively, an update function, a filter function, or a validate_doc_updates function.
		- func_name: The name of the function. Ignored for validate functions.
		- func_body: The function body.
		- auto_update: A boolean indicating whether the view should be automatically updated when the design document is updated. Defaults to true.
	*/
	`CREATE TABLE {{ .Design }} (
		id TEXT NOT NULL,
		rev INTEGER NOT NULL,
		rev_id TEXT NOT NULL,
		language TEXT NOT NULL DEFAULT 'javascript',
		func_type TEXT CHECK (func_type IN ('map', 'reduce', 'update', 'filter', 'validate')) NOT NULL,
		func_name TEXT NOT NULL,
		func_body TEXT NOT NULL,
		auto_update BOOLEAN NOT NULL DEFAULT TRUE,
		-- Options include_design and local_seq are only stored for 'map' type
		include_design BOOLEAN,
		local_seq BOOLEAN,
		last_seq INTEGER, -- the last map-indexed sequence id, NULL for others
		FOREIGN KEY (id, rev, rev_id) REFERENCES {{ .Docs }} (id, rev, rev_id) ON DELETE CASCADE,
		UNIQUE (id, rev, rev_id, func_type, func_name)
	)`,
}

var viewSchema = []string{
	`CREATE TABLE {{ .Map }} (
		pk INTEGER PRIMARY KEY,
		id TEXT NOT NULL,
		rev INTEGER NOT NULL,
		rev_id TEXT NOT NULL,
		key TEXT COLLATE COUCHDB_UCI,
		value TEXT,
		FOREIGN KEY (id, rev, rev_id) REFERENCES {{ .Docs }} (id, rev, rev_id)
	)`,
	`CREATE INDEX {{ .IndexMap }} ON {{ .Map }} (key)`,
}
