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

//go:build js

package internal

import (
	"context"
	"testing"

	"github.com/go-kivik/kivik/v4"
)

var pouchdbVer *string

// PouchDBVersion returns the version of PouchDB library being used.
func PouchDBVersion(t *testing.T) string {
	t.Helper()

	if pouchdbVer != nil {
		return *pouchdbVer
	}
	client, err := kivik.New("pouch", "")
	if err != nil {
		t.Fatal(err)
	}
	v, err := client.Version(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	pouchdbVer = &v.Version
	return v.Version
}

// MustPouchDBVersion returns the version of PouchDB library being used.
func MustPouchDBVersion() string {
	if pouchdbVer != nil {
		return *pouchdbVer
	}
	client, err := kivik.New("pouch", "")
	if err != nil {
		panic(err)
	}
	v, err := client.Version(context.Background())
	if err != nil {
		panic(err)
	}
	pouchdbVer = &v.Version
	return v.Version
}
