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

import "github.com/go-kivik/kivik/v4/driver"

type optsMap map[string]interface{}

func newOpts(options driver.Options) optsMap {
	opts := map[string]interface{}{}
	options.Apply(opts)
	return opts
}

func (o optsMap) endKey() string {
	if endkey, ok := o["endkey"].(string); ok {
		return endkey
	}
	if endkey, ok := o["end_key"].(string); ok {
		return endkey
	}
	return ""
}

func (o optsMap) inclusiveEnd() bool {
	inclusiveEnd, ok := o["inclusive_end"].(bool)
	return !ok || inclusiveEnd
}

func (o optsMap) startKey() string {
	if startkey, ok := o["startkey"].(string); ok {
		return startkey
	}
	if startkey, ok := o["start_key"].(string); ok {
		return startkey
	}
	return ""
}

func (o optsMap) rev() string {
	rev, _ := o["rev"].(string)
	return rev
}
