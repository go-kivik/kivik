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

// Package cookies provides cookies utilities.
package cookies

import (
	"bytes"
	"encoding/base64"
	"strconv"

	"github.com/pkg/errors"
)

// DecodeCookie decodes a Base64-encoded cookie, and returns its component
// parts.
func DecodeCookie(cookie string) (name string, created int64, err error) {
	data, err := base64.RawURLEncoding.DecodeString(cookie)
	if err != nil {
		return "", 0, err
	}
	const partCount = 3
	parts := bytes.SplitN(data, []byte(":"), partCount)
	t, err := strconv.ParseInt(string(parts[1]), 16, 64)
	if err != nil {
		return "", 0, errors.Wrap(err, "invalid timestamp")
	}
	return string(parts[0]), t, nil
}
