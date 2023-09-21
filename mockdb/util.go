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

package mockdb

import (
	"fmt"
	"time"

	kivik "github.com/go-kivik/kivik/v4"
)

func optionsString(opt kivik.Option) string {
	if opt == nil {
		return "\n\t- has any options"
	}
	return fmt.Sprintf("\n\t- has options: %s", opt)
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return fmt.Sprintf("\n\t- should return error: %s", err)
}

func delayString(delay time.Duration) string {
	if delay == 0 {
		return ""
	}
	return fmt.Sprintf("\n\t- should delay for: %s", delay)
}

func fieldString(field, value string) string {
	if value == "" {
		return "\n\t- has any " + field
	}
	return "\n\t- has " + field + ": " + value
}
