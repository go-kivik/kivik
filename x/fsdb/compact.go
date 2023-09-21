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
	"context"
	"strings"

	"github.com/go-kivik/fsdb/v4/cdb"
	"github.com/go-kivik/fsdb/v4/cdb/decode"
	"github.com/go-kivik/fsdb/v4/filesystem"
	"github.com/go-kivik/kivik/v4"
)

type docIndex map[string]*cdb.Document

func (i docIndex) readIndex(ctx context.Context, fs filesystem.Filesystem, path string) error {
	dir, err := fs.Open(path)
	if err != nil {
		return kerr(err)
	}
	files, err := dir.Readdir(-1)
	if err != nil {
		return kerr(err)
	}

	c := cdb.New(path, fs)

	var docID string
	for _, info := range files {
		if err := ctx.Err(); err != nil {
			return err
		}
		switch {
		case !info.IsDir():
			id, _, ok := decode.ExplodeFilename(info.Name())
			if !ok {
				// ignore unrecognized files
				continue
			}
			docID = id
		case info.IsDir() && info.Name()[0] == '.':
			docID = strings.TrimPrefix(info.Name(), ".")
		default:
			continue
		}
		if _, ok := i[docID]; ok {
			// We've already read this one
			continue
		}
		doc, err := c.OpenDocID(docID, kivik.Params(nil))
		if err != nil {
			return err
		}
		i[docID] = doc
	}
	return nil
}

func (d *db) Compact(ctx context.Context) error {
	return d.compact(ctx, filesystem.Default())
}

func (d *db) compact(ctx context.Context, fs filesystem.Filesystem) error {
	docs := docIndex{}
	if err := docs.readIndex(ctx, fs, d.path()); err != nil {
		return err
	}
	for _, doc := range docs {
		if err := doc.Compact(ctx); err != nil {
			return err
		}
	}
	return nil
}
