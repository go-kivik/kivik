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

package fs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/otiai10/copy"
)

func tempDir(t *testing.T) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "kivik-fsdb-")
	if err != nil {
		t.Fatal(err)
	}
	return dir
}

func rmdir(t *testing.T, dir string) {
	t.Helper()
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}

// copyDir recursively copies the contents of path to a new temporary dir, whose
// path is returned. The depth argument controls how deeply path is placed into
// the temp dir. Examples:
//
//	copyDir(t, "/foo/bar/baz", 0) // copies /foo/bar/baz/* to /tmp-XXX/*
//	copyDir(t, "/foo/bar/baz", 1) // copies /foo/bar/baz/* to /tmp-XXX/baz/*
//	copyDir(t, "/foo/bar/baz", 3) // copies /foo/bar/baz/* to /tmp-XXX/foo/bar/baz/*
func copyDir(t *testing.T, source string, depth int) string { // nolint: unparam
	t.Helper()
	tmpdir := tempDir(t)
	target := tmpdir
	if depth > 0 {
		parts := strings.Split(source, string(filepath.Separator))
		if len(parts) < depth {
			t.Fatalf("Depth of %d specified, but path only has %d parts", depth, len(parts))
		}
		target = filepath.Join(append([]string{tmpdir}, parts[len(parts)-depth:]...)...)
		if err := os.MkdirAll(target, 0o777); err != nil {
			t.Fatal(err)
		}
	}
	if err := copy.Copy(source, target); err != nil {
		t.Fatal(err)
	}
	return tmpdir
}

func cleanTmpdir(path string) func() error {
	return func() error {
		return os.RemoveAll(path)
	}
}
