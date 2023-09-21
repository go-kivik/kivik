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

package cdb

import (
	"context"
	"os"
	"path/filepath"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/fsdb/cdb/decode"
)

// ReadSecurity reads the _security.{ext} document from path.
func (fs *FS) ReadSecurity(ctx context.Context, path string) (*driver.Security, error) {
	sec := new(driver.Security)
	for _, ext := range decode.Extensions() {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		f, err := fs.fs.Open(filepath.Join(path, "_security."+ext))
		if err == nil {
			defer f.Close() // nolint: errcheck
			err := decode.Decode(f, ext, sec)
			return sec, err
		}
		if !os.IsNotExist(err) {
			return nil, err
		}
	}
	return sec, nil
}
