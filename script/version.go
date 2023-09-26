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

// Package main prints out the version constant, for use in automatically
// creating releases when the version is updated.
package main

import (
	"fmt"
	"strings"

	"github.com/go-kivik/kivik/v4"
)

func main() {
	if strings.HasSuffix(kivik.Version, "-prerelease") {
		return
	}
	fmt.Printf("TAG=v%s\n", kivik.Version)
	if strings.Contains(kivik.Version, "-") {
		fmt.Println("PRERELEASE=true")
	}
}
