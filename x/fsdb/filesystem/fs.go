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

// Package filesystem provides an abstraction around a filesystem
package filesystem

import (
	"io"
	"os"
)

// Filesystem is a filesystem implementation.
type Filesystem interface {
	Mkdir(name string, perm os.FileMode) error
	MkdirAll(path string, perm os.FileMode) error
	Open(string) (File, error)
	Create(string) (File, error)
	Stat(string) (os.FileInfo, error)
	TempFile(dir, pattern string) (File, error)
	Rename(oldpath, newpath string) error
	Remove(name string) error
	Link(oldname, newname string) error
}

type defaultFS struct{}

var _ Filesystem = &defaultFS{}

func (fs *defaultFS) Mkdir(name string, perm os.FileMode) error {
	return os.Mkdir(name, perm)
}

func (fs *defaultFS) MkdirAll(path string, perm os.FileMode) error {
	return os.MkdirAll(path, perm)
}

func (fs *defaultFS) Open(name string) (File, error) {
	return os.Open(name)
}

func (fs *defaultFS) TempFile(dir, pattern string) (File, error) {
	return os.CreateTemp(dir, pattern)
}

func (fs *defaultFS) Create(name string) (File, error) {
	return os.Create(name)
}

func (fs *defaultFS) Stat(name string) (os.FileInfo, error) {
	return os.Stat(name)
}

func (fs *defaultFS) Rename(oldpath, newpath string) error {
	return os.Rename(oldpath, newpath)
}

func (fs *defaultFS) Remove(name string) error {
	return os.Remove(name)
}

func (fs *defaultFS) Link(oldname, newname string) error {
	return os.Link(oldname, newname)
}

// File represents a file object.
type File interface {
	io.Reader
	io.Closer
	io.Writer
	io.ReaderAt
	io.Seeker
	Name() string
	Readdir(int) ([]os.FileInfo, error)
	Stat() (os.FileInfo, error)
}

type defaultFile struct {
	*os.File
}

var _ File = &defaultFile{}

// Default returns the default filesystem implementation.
func Default() Filesystem {
	return &defaultFS{}
}
