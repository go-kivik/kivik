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
	"encoding/json"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/fsdb/cdb/decode"
)

// ReadSecurity reads the _security.{ext} document from path. If not found,
// an empty security document is returned.
func (fs *FS) ReadSecurity(ctx context.Context, path string) (*driver.Security, error) {
	sec := new(driver.Security)
	ext, err := fs.findSecurityExt(ctx, path)
	if err != nil {
		return nil, err
	}
	if ext == "" {
		return sec, nil
	}
	f, err := fs.fs.Open(filepath.Join(path, "_security."+ext))
	if err == nil {
		defer f.Close()
		err := decode.Decode(f, ext, sec)
		return sec, err
	}
	if !os.IsNotExist(err) {
		return nil, err
	}

	return sec, nil
}

// findSecurityExt looks in path for the `_security` document by extension, and
// returns the first matching extension it finds, or an empty string if the
// security document is not found.
func (fs *FS) findSecurityExt(ctx context.Context, path string) (string, error) {
	for _, ext := range decode.Extensions() {
		if err := ctx.Err(); err != nil {
			return "", err
		}
		stat, err := fs.fs.Stat(filepath.Join(path, "_security."+ext))
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		if stat.Mode().IsRegular() {
			return ext, nil
		}
	}
	return "", nil
}

// WriteSecurity overwrites the existing _security document, if it exists. If it
// does not exist, it is created with the .json extension.
func (fs *FS) WriteSecurity(ctx context.Context, path string, sec *driver.Security) error {
	ext, err := fs.findSecurityExt(ctx, path)
	if err != nil {
		return err
	}
	if ext == "" {
		ext = "json"
	}
	filename := filepath.Join(path, "_security."+ext)
	tmp, err := fs.fs.TempFile(path, "_security."+ext)
	if err != nil {
		return err
	}
	var enc interface {
		Encode(interface{}) error
	}
	switch ext {
	case "json":
		enc = json.NewEncoder(tmp)
	case "yaml", "yml":
		enc = yaml.NewEncoder(tmp)
	}
	if err := enc.Encode(sec); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return fs.fs.Rename(tmp.Name(), filename)
}
