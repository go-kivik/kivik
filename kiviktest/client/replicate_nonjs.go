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

package client

import (
	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func replicationOptions(_ *kt.Context, _ *kivik.Client, _, _, _ string, in map[interface{}]interface{}) map[interface{}]interface{} {
	if in == nil {
		in = make(map[interface{}]interface{})
	}
	return in
}
