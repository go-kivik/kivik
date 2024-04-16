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

package sqlite

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"
	"text/template"
)

// The global template cache.
var (
	mu        sync.Mutex
	tmplCache = template.New("")
)

type tmplFuncs struct {
	db *db
}

func (t *tmplFuncs) Docs() string {
	return strconv.Quote(t.db.name)
}

func (t *tmplFuncs) Revs() string {
	return strconv.Quote(t.db.name + "_revs")
}

func (t *tmplFuncs) Attachments() string {
	return strconv.Quote(t.db.name + "_attachments")
}

func (t *tmplFuncs) AttachmentsBridge() string {
	return strconv.Quote(t.db.name + "_attachments_bridge")
}

func (t *tmplFuncs) Leaves() string {
	return strconv.Quote(t.db.name + "_leaves")
}

func (t *tmplFuncs) Design() string {
	return strconv.Quote(t.db.name + "_design")
}

// query does variable substitution on a query string. The following transitions
// are made:
//
//	{{ .Docs }} -> db.name
//	{{ .Revs }} -> db.name + "_revs"
//	{{ .Attachments }} -> db.name + "_attachments"
//	{{ .AttachmentsBridge }} -> db.name + "_attachments_bridge"
//	{{ .Leaves }} -> db.name + "_leaves"
//	{{ .Design }} -> db.name + "_design"
func (d *db) query(format string) string {
	var buf bytes.Buffer
	tmpl := getTmpl(format)

	if err := tmpl.Execute(&buf, &tmplFuncs{db: d}); err != nil {
		panic(err)
	}
	return buf.String()
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
	return fmt.Sprintf("%x", hash.Sum(nil))
}
