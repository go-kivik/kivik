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

// Package main generates the bulk of the mockdb driver.
package main

import (
	"os"
	"reflect"
	"sort"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

var clientSkips = map[string]struct{}{
	"Driver":   {},
	"DSN":      {},
	"CreateDB": {},
}

var dbSkips = map[string]struct{}{
	"Close":         {},
	"Client":        {},
	"Err":           {},
	"Name":          {},
	"Search":        {},
	"SearchAnalyze": {},
	"SearchInfo":    {},
}

func main() {
	initTemplates(os.Args[1])
	const openPerms = 0o777
	if err := os.Mkdir("./other", openPerms); err != nil && !os.IsExist(err) {
		panic(err)
	}
	if err := client(); err != nil {
		panic(err)
	}
	if err := db(); err != nil {
		panic(err)
	}
}

type fullClient interface {
	driver.Client
	driver.DBsStatser
	driver.Pinger
	driver.Sessioner
	driver.Cluster
	driver.ClientCloser
	driver.ClientReplicator
	driver.DBUpdater
	driver.Configer
}

func client() error {
	dMethods, err := parseMethods(struct{ X fullClient }{}, false, clientSkips) // nolint: unused
	if err != nil {
		return err
	}

	client, err := parseMethods(struct{ X *kivik.Client }{}, true, clientSkips) // nolint: unused
	if err != nil {
		return err
	}
	same, cm, dm := compareMethods(client, dMethods)

	if err := renderExpectationsGo("clientexpectations_gen.go", append(same, dm...)); err != nil {
		return err
	}
	if err := renderClientGo("client_gen.go", append(same, dm...)); err != nil {
		return err
	}
	return renderMockGo("clientmock_gen.go", append(same, cm...))
}

type fullDB interface {
	driver.DB
	driver.AttachmentMetaGetter
	driver.BulkDocer
	driver.BulkGetter
	driver.Copier
	driver.DesignDocer
	driver.Finder
	driver.Flusher
	driver.LocalDocer
	driver.RevGetter
	driver.Purger
	driver.RevsDiffer
	driver.PartitionedDB
	driver.SecurityDB
	driver.RowsGetter
}

func db() error {
	dMethods, err := parseMethods(struct{ X fullDB }{}, false, dbSkips) // nolint: unused
	if err != nil {
		return err
	}

	client, err := parseMethods(struct{ X *kivik.DB }{}, true, dbSkips) // nolint: unused
	if err != nil {
		return err
	}
	same, cm, dm := compareMethods(client, dMethods)

	for _, method := range same {
		method.DBMethod = true
	}
	for _, method := range dm {
		method.DBMethod = true
	}
	for _, method := range cm {
		method.DBMethod = true
	}

	if err := renderExpectationsGo("dbexpectations_gen.go", append(same, dm...)); err != nil {
		return err
	}
	if err := renderClientGo("db_gen.go", append(same, dm...)); err != nil {
		return err
	}
	return renderMockGo("dbmock_gen.go", append(same, cm...))
}

func compareMethods(client, driver []*method) (same []*method, differentClient []*method, differentDriver []*method) {
	dMethods := toMap(driver)
	cMethods := toMap(client)
	sameMethods := make(map[string]*method)
	for name, method := range dMethods {
		if cMethod, ok := cMethods[name]; ok {
			if reflect.DeepEqual(cMethod, method) {
				sameMethods[name] = method
				delete(dMethods, name)
				delete(cMethods, name)
			}
		} else {
			delete(dMethods, name)
			delete(cMethods, name)
		}
	}
	return toSlice(sameMethods), toSlice(cMethods), toSlice(dMethods)
}

func toSlice(methods map[string]*method) []*method {
	result := make([]*method, 0, len(methods))
	for _, method := range methods {
		result = append(result, method)
	}
	sort.Slice(result, func(i, j int) bool {
		return result[i].Name < result[j].Name
	})
	return result
}

func toMap(methods []*method) map[string]*method {
	result := make(map[string]*method, len(methods))
	for _, method := range methods {
		result[method.Name] = method
	}
	return result
}
