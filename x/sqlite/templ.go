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

package sqlite

import (
	"bytes"
	"encoding/hex"
	"hash/fnv"
	"strconv"
	"strings"
	"sync"
	"text/template"
)

// The global template cache.
var (
	mu        sync.Mutex
	tmplCache = template.New("")
)

type tmplFuncs struct {
	db                  *db
	ddoc, viewName, rev string
	hash                string
	collation           *string
}

const tablePrefix = "kivik$"

func (t *tmplFuncs) tableName(suffix string) string {
	return strconv.Quote(tablePrefix + t.db.name + suffix)
}

func (t *tmplFuncs) indexName(suffix string) string {
	return strconv.Quote("idx_" + tablePrefix + t.db.name + suffix)
}

func (t *tmplFuncs) Docs() string              { return t.tableName("") }
func (t *tmplFuncs) Revs() string              { return t.tableName("$revs") }
func (t *tmplFuncs) Attachments() string       { return t.tableName("$attachments") }
func (t *tmplFuncs) AttachmentsBridge() string { return t.tableName("$attachments_bridge") }
func (t *tmplFuncs) Design() string            { return t.tableName("$design") }
func (t *tmplFuncs) Security() string          { return t.tableName("$security") }
func (t *tmplFuncs) MangoIndexes() string      { return t.tableName("$mango_indexes") }
func (t *tmplFuncs) IndexKey() string          { return t.indexName("$key") }
func (t *tmplFuncs) IndexParent() string       { return t.indexName("$parent") }

const maxTableLen = 59 // 64 minus the `idx_` prefix, and one more `_` separator

// hashedName returns a table name in the format "{{db name}}_{{ddoc}}_{{typ}}_{{hash}}"
// where hash is the first 8 characters of the MD5 sum of the dbname, ddoc, and type.
// If the final version is longer than 64 characters, it is truncated to size,
// before appending the hash.
func (t *tmplFuncs) hashedName(typ string) string {
	if t.ddoc == "" {
		panic("ddoc template method called outside of a ddoc template")
	}
	name := strings.Join([]string{t.ddoc, t.rev, t.viewName}, "_")
	if t.hash == "" {
		t.hash = md5sumString(name)[:8]
	}
	name += "_" + typ
	if len(name) > maxTableLen-len(t.hash) {
		name = name[:maxTableLen-len(t.hash)]
	}
	return name + "_" + t.hash
}

func (t *tmplFuncs) Map() string {
	return strconv.Quote(tablePrefix + t.hashedName("map"))
}

func (t *tmplFuncs) IndexMap() string {
	return strconv.Quote("idx_" + t.hashedName("map"))
}

func (t *tmplFuncs) Collation() string {
	if t.collation == nil {
		return "COUCHDB_UCI"
	}
	switch *t.collation {
	case "ascii", "raw":
		return "BINARY"
	default:
		panic("unsupported collation: " + *t.collation)
	}
}

func executeTmpl(format string, data *tmplFuncs) string {
	var buf bytes.Buffer
	tmpl := getTmpl(format)

	if err := tmpl.Execute(&buf, data); err != nil {
		panic(err)
	}
	return buf.String()
}

// query does variable substitution on a query string. The following translations
// are made:
//
//	{{ .Docs }}              -> "kivik$" + db.name
//	{{ .Revs }}              -> "kivik$" + db.name + "$revs"
//	{{ .Attachments }}       -> "kivik$" + db.name + "$attachments"
//	{{ .AttachmentsBridge }} -> "kivik$" + db.name + "$attachments_bridge"
//	{{ .Design }}            -> "kivik$" + db.name + "$design"
func (d *db) query(format string) string {
	return executeTmpl(format, &tmplFuncs{db: d})
}

// ddocQuery works just like [db.query], but also enables access to the
// following translations:
//
//	{{ .Map }}      -> the view map table name
//	{{ .IndexMap }} -> the view map index name
func (d *db) ddocQuery(docID, viewOrFuncName, rev, format string) string {
	return executeTmpl(format, &tmplFuncs{
		db:       d,
		ddoc:     strings.TrimPrefix(docID, "_design/"),
		viewName: viewOrFuncName,
		rev:      rev,
	})
}

// createDdocQuery works just like [db.ddocQuery], but also enables access to the
// following translations:
//
//	{{ .Collation }} -> the view's collation sequence
func (d *db) createDdocQuery(docID, viewOrFuncName, rev, format string, collation *string) string {
	return executeTmpl(format, &tmplFuncs{
		db:        d,
		ddoc:      strings.TrimPrefix(docID, "_design/"),
		viewName:  viewOrFuncName,
		rev:       rev,
		collation: collation,
	})
}

func getTmpl(format string) *template.Template {
	mu.Lock()
	defer mu.Unlock()
	name := calcTmplName(format)
	if tmpl := tmplCache.Lookup(name); tmpl != nil {
		return tmpl
	}
	newTmpl := tmplCache.New(name)
	_, err := newTmpl.Parse(format)
	if err != nil {
		panic(err)
	}
	return newTmpl
}

func calcTmplName(format string) string {
	hash := fnv.New128a()
	hash.Write([]byte(format))
	return hex.EncodeToString(hash.Sum(nil))
}
