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

package memorydb

import "fmt"

func randStr() string {
	rndMU.Lock()
	s := fmt.Sprintf("%016x%016x", rnd.Uint64(), rnd.Uint64())
	rndMU.Unlock()
	return s
}
