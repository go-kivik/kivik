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

package ast

// Match returns true if the selector matches the input document. doc is
// expected to be the result of unmarshaling JSON to an empty interface. An
// invalid document will cause Match to panic.
func Match(sel Selector, doc interface{}) bool {
	if sel == nil {
		return true
	}
	m, ok := sel.(interface{ Match(interface{}) bool })
	if !ok {
		return false
	}
	return m.Match(doc)
}
