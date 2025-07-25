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

// Package decode assists in document decoding.
package decode

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

type decoder interface {
	Decode(io.Reader, interface{}) error
}

var decoders = map[string]decoder{
	"json": &jsonDecoder{},
	"yaml": &yamlDecoder{},
	"yml":  &yamlDecoder{},
}

var extensions = func() []string {
	exts := make([]string, 0, len(decoders))
	for ext := range decoders {
		exts = append(exts, ext)
	}
	sort.Strings(exts)
	return exts
}()

// OpenAny attempts to open base + any supported extension. It returns the open
// file, the matched extension, or an error.
func OpenAny(fs filesystem.Filesystem, base string) (f filesystem.File, ext string, err error) {
	for ext = range decoders {
		f, err = fs.Open(base + "." + ext)
		if err == nil || !os.IsNotExist(err) {
			return
		}
	}
	return
}

// Decode decodes r according to ext's registered decoder, into i.
func Decode(r io.Reader, ext string, i interface{}) error {
	ext = strings.TrimPrefix(ext, ".")
	dec, ok := decoders[ext]
	if !ok {
		return fmt.Errorf("no decoder for %s", ext)
	}
	return dec.Decode(r, i)
}

// ExplodeFilename returns the base name, extension, and a boolean indicating
// whether the extension is decodable.
func ExplodeFilename(filename string) (basename, ext string, ok bool) {
	dotExt := filepath.Ext(filename)
	basename = strings.TrimSuffix(filename, dotExt)
	ext = strings.TrimPrefix(dotExt, ".")
	_, ok = decoders[ext]
	return basename, ext, ok
}

// Extensions returns a list of supported extensions.
func Extensions() []string {
	return extensions
}
