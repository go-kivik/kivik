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

package client

import (
	"testing"

	"github.com/gopherjs/gopherjs/js"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

type multiOptions []kivik.Option

func (mo multiOptions) Apply(target any) {
	for _, o := range mo {
		if o != nil {
			o.Apply(target)
		}
	}
}

func replicationOptions(t *testing.T, c *kt.Context, target, source, repID string, in kivik.Option) kivik.Option {
	if c.String(t, "mode") != "pouchdb" {
		return multiOptions{kivik.Param("_id", repID), in}
	}
	return multiOptions{
		kivik.Param("source", js.Global.Get("PouchDB").New(source)),
		kivik.Param("target", js.Global.Get("PouchDB").New(target)),
		in,
	}
}
