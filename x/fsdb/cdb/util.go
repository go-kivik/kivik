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
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"io"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

var reservedKeys = map[string]struct{}{
	"_id":                {},
	"_rev":               {},
	"_attachments":       {},
	"_revisions":         {},
	"_revs_info":         {}, // *
	"_deleted":           {},
	"_conflicts":         {}, // *
	"_deleted_conflicts": {}, // *
	"_local_seq":         {}, // *
	// * Can these be PUT?
}

// EscapeID escapes a document or database ID to be a safe filename.
func EscapeID(s string) string {
	if s == "" {
		return s
	}
	if strings.Contains(s, "/") || strings.Contains(s, "%2F") || strings.Contains(s, "%2f") {
		s = strings.ReplaceAll(s, "%", "%25")
		s = strings.ReplaceAll(s, "/", "%2F")
	}
	return s
}

// UnescapeID unescapes a filename into a document or database ID.
func UnescapeID(s string) string {
	if s == "" {
		return s
	}
	if strings.Contains(s, "%2F") || strings.Contains(s, "%2f") || strings.Contains(s, "%25") {
		s = strings.ReplaceAll(s, "%2F", "/")
		s = strings.ReplaceAll(s, "%2f", "/")
		s = strings.ReplaceAll(s, "%25", "%")
	}
	return s
}

// copyDigest works the same as io.Copy, but also returns the md5sum
// digest of the copied file.
func copyDigest(tgt io.Writer, dst io.Reader) (int64, string, error) {
	h := md5.New()
	tee := io.TeeReader(dst, h)
	wg := sync.WaitGroup{}
	wg.Add(1)
	var written int64
	var err error
	go func() {
		written, err = io.Copy(tgt, tee)
		wg.Done()
	}()
	wg.Wait()

	return written, "md5-" + base64.StdEncoding.EncodeToString(h.Sum(nil)), err
}

func digest(r io.Reader) (int64, string) {
	h := md5.New()
	written, _ := io.Copy(h, r)
	return written, "md5-" + base64.StdEncoding.EncodeToString(h.Sum(nil))
}

func joinJSON(objects ...json.RawMessage) []byte {
	var size int
	for _, obj := range objects {
		size += len(obj)
	}
	result := make([]byte, 0, size)
	result = append(result, '{')
	for _, obj := range objects {
		if len(obj) == 4 && string(obj) == "null" {
			continue
		}
		result = append(result, obj[1:len(obj)-1]...)
		result = append(result, ',')
	}
	result[len(result)-1] = '}'
	return result
}

func atomicWriteFile(fs filesystem.Filesystem, path string, r io.Reader) error {
	f, err := fs.TempFile(filepath.Dir(path), ".tmp."+filepath.Base(path)+"-")
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, r); err != nil {
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return fs.Rename(f.Name(), path)
}

type atomicWriter struct {
	fs   filesystem.Filesystem
	path string
	f    filesystem.File
	err  error
}

func (w *atomicWriter) Write(p []byte) (int, error) {
	if w.err != nil {
		return 0, w.err
	}
	return w.f.Write(p)
}

func (w *atomicWriter) Close() error {
	if w.err != nil {
		return w.err
	}
	if err := w.f.Close(); err != nil {
		return err
	}
	return w.fs.Rename(w.f.Name(), w.path)
}

// atomicFileWriter returns an io.WriteCloser, which writes to a temp file, then
// when Close() is called, it renames to the originally requested path.
func atomicFileWriter(fs filesystem.Filesystem, path string) io.WriteCloser {
	f, err := fs.TempFile(filepath.Dir(path), ".tmp."+filepath.Base(path)+"-")
	return &atomicWriter{
		fs:   fs,
		path: path,
		f:    f,
		err:  err,
	}
}
