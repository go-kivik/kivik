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

package bindings

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

// RecoverError recovers from a thrown JS error. If an error is caught, err
// is set to its value.
//
// To use, put this at the beginning of a function:
//
//	defer RecoverError(&err)
func RecoverError(err *error) {
	if r := recover(); r != nil {
		switch t := r.(type) {
		case *js.Object:
			*err = NewPouchError(t)
		case error:
			// This shouldn't ever happen, but just in case
			*err = t
		default:
			// Catch all for everything else
			*err = fmt.Errorf("%v", r)
		}
	}
}
