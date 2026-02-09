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
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-kivik/kivik/v4/driver"
	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

// Document is a CouchDB document.
type Document struct {
	ID        string    `json:"_id" yaml:"_id"`
	Revisions Revisions `json:"-" yaml:"-"`
	// RevsInfo is only used during JSON marshaling when revs_info=true, and
	// should never be consulted as authoritative.
	RevsInfo []RevInfo `json:"_revs_info,omitempty" yaml:"-"`
	// RevHistory is only used during JSON marshaling, when revs=true, and
	// should never be consulted as authoritative.
	RevHistory *RevHistory `json:"_revisions,omitempty" yaml:"-"`

	Options map[string]any `json:"-" yaml:"-"`

	cdb *FS
}

// NewDocument creates a new document.
func (fs *FS) NewDocument(docID string) *Document {
	return &Document{
		ID:  docID,
		cdb: fs,
	}
}

// MarshalJSON satisfies the json.Marshaler interface.
func (d *Document) MarshalJSON() ([]byte, error) {
	d.revsInfo()
	d.revs()
	rev := d.Revisions[0]
	rev.options = d.Options
	revJSON, err := json.Marshal(rev)
	if err != nil {
		return nil, err
	}
	docJSON, _ := json.Marshal(*d)
	return joinJSON(docJSON, revJSON), nil
}

// revs populates the Rev
func (d *Document) revs() {
	d.RevHistory = nil
	if ok, _ := d.Options["revs"].(bool); !ok {
		return
	}
	if len(d.Revisions) < 1 {
		return
	}
	d.RevHistory = d.Revisions[0].RevHistory
}

// revsInfo populates the RevsInfo field, if appropriate according to options.
func (d *Document) revsInfo() {
	d.RevsInfo = nil
	if ok, _ := d.Options["revs_info"].(bool); !ok {
		return
	}
	if _, ok := d.Options["rev"]; ok {
		return
	}
	d.RevsInfo = make([]RevInfo, len(d.Revisions))
	for i, rev := range d.Revisions {
		d.RevsInfo[i] = RevInfo{
			Rev:    rev.Rev.String(),
			Status: "available",
		}
	}
}

// RevInfo is revisions information as presented in the _revs_info key.
type RevInfo struct {
	Rev    string `json:"rev"`
	Status string `json:"status"`
}

// Compact cleans up any non-leaf revs, and attempts to consolidate attachments.
func (d *Document) Compact(ctx context.Context) error {
	revTree := make(map[string]*Revision, 1)
	// An index of ancestor -> leaf revision
	index := map[string][]string{}
	keep := make([]*Revision, 0, 1)
	for _, rev := range d.Revisions {
		revID := rev.Rev.String()
		if leafIDs, ok := index[revID]; ok {
			for _, leafID := range leafIDs {
				if err := copyAttachments(d.cdb.fs, revTree[leafID], rev); err != nil {
					return err
				}
			}
			if err := rev.Delete(ctx); err != nil {
				return err
			}
			continue
		}
		keep = append(keep, rev)
		for _, ancestor := range rev.RevHistory.Ancestors()[1:] {
			index[ancestor] = append(index[ancestor], revID)
		}
		revTree[revID] = rev
	}
	d.Revisions = keep
	return nil
}

const tempPerms = 0o777

func copyAttachments(fs filesystem.Filesystem, leaf, old *Revision) error {
	leafpath := strings.TrimSuffix(leaf.path, filepath.Ext(leaf.path)) + "/"
	basepath := strings.TrimSuffix(old.path, filepath.Ext(old.path)) + "/"
	for filename, att := range old.Attachments {
		if _, ok := leaf.Attachments[filename]; !ok {
			continue
		}
		if strings.HasPrefix(att.path, basepath) {
			name := filepath.Base(att.path)
			if err := os.MkdirAll(leafpath, tempPerms); err != nil {
				return err
			}
			if err := fs.Link(att.path, filepath.Join(leafpath, name)); err != nil {
				if os.IsExist(err) {
					if err := fs.Remove(att.path); err != nil {
						return err
					}
					continue
				}
				return err
			}
		}
	}
	return nil
}

// AddRevision adds rev to the existing document, according to options, and
// persists it to disk. The return value is the new revision ID.
func (d *Document) AddRevision(ctx context.Context, rev *Revision, options driver.Options) (string, error) {
	revid, err := d.addRevision(ctx, rev, options)
	if err != nil {
		return "", err
	}
	err = d.persist(ctx)
	return revid, err
}

func (d *Document) addOldEdit(rev *Revision) (string, error) {
	if rev.Rev.IsZero() {
		return "", statusError{status: http.StatusBadRequest, error: errors.New("_rev required with new_edits=false")}
	}
	for _, r := range d.Revisions {
		if r.Rev.Equal(rev.Rev) {
			// If the rev already exists, do nothing, but report success.
			return r.Rev.String(), nil
		}
	}
	d.Revisions = append(d.Revisions, rev)
	sort.Sort(d.Revisions)
	return rev.Rev.String(), nil
}

func (d *Document) addRevision(ctx context.Context, rev *Revision, options driver.Options) (string, error) {
	opts := map[string]any{}
	options.Apply(opts)
	if newEdits, ok := opts["new_edits"].(bool); ok && !newEdits {
		return d.addOldEdit(rev)
	}
	if revid, ok := opts["rev"].(string); ok {
		var newrev RevID
		if err := newrev.UnmarshalText([]byte(revid)); err != nil {
			return "", err
		}
		if !rev.Rev.IsZero() && rev.Rev.String() != newrev.String() {
			return "", statusError{status: http.StatusBadRequest, error: errors.New("document rev from request body and query string have different values")}
		}
		rev.Rev = newrev
	}
	needRev := len(d.Revisions) > 0
	haveRev := !rev.Rev.IsZero()
	if needRev != haveRev {
		return "", errConflict
	}
	var oldrev *Revision
	if len(d.Revisions) > 0 {
		var ok bool
		if oldrev, ok = d.leaves()[rev.Rev.String()]; !ok {
			return "", errConflict
		}
	}

	hash, err := rev.hash()
	if err != nil {
		return "", err
	}
	rev.Rev = RevID{
		Seq: rev.Rev.Seq + 1,
		Sum: hash,
	}
	if oldrev != nil {
		rev.RevHistory = oldrev.RevHistory.AddRevision(rev.Rev)
	}

	revpath := filepath.Join(d.cdb.root, "."+EscapeID(d.ID), rev.Rev.String())
	var dirMade bool
	for filename, att := range rev.Attachments {
		att.fs = d.cdb.fs
		if err := ctx.Err(); err != nil {
			return "", err
		}
		var oldatt *Attachment
		if oldrev != nil {
			oldatt = oldrev.Attachments[filename]
		}
		if !att.Stub {
			revpos := rev.Rev.Seq
			att.RevPos = &revpos
			if !dirMade {
				if err := d.cdb.fs.MkdirAll(revpath, tempPerms); err != nil && !os.IsExist(err) {
					return "", err
				}
				dirMade = true
			}
			if err := att.persist(revpath, filename); err != nil {
				return "", err
			}
			if oldatt != nil && oldatt.Digest == att.Digest {
				if err := att.fs.Remove(att.path); err != nil {
					return "", err
				}
				att.path = ""
				att.Stub = true
				att.RevPos = oldatt.RevPos
			}
			continue
		}
		if oldrev == nil {
			// Can't upload stubs if there's no previous revision
			return "", statusError{status: http.StatusInternalServerError, error: fmt.Errorf("attachment %s: %w", filename, err)}
		}
		if att.Digest != "" && att.Digest != oldatt.Digest {
			return "", statusError{status: http.StatusBadRequest, error: fmt.Errorf("invalid attachment data for %s", filename)}
		}
		if att.RevPos != nil && *att.RevPos != *oldatt.RevPos {
			return "", statusError{status: http.StatusBadRequest, error: fmt.Errorf("invalid attachment data for %s", filename)}
		}
	}

	if len(d.Revisions) == 0 {
		rev.RevHistory = &RevHistory{
			Start: rev.Rev.Seq,
			IDs:   []string{rev.Rev.Sum},
		}
	}
	d.Revisions = append(d.Revisions, rev)
	sort.Sort(d.Revisions)
	return rev.Rev.String(), nil
}

/*
	persist updates the current rev state on disk.

Persist strategy:

  - For every rev that doesn't exist on disk, create it in {db}/.{docid}/{rev}
  - If winning rev does not exist in {db}/{docid}:
  - Move old winning rev to {db}/.{docid}/{rev}
  - Move new winning rev to {db}/{docid}
*/
func (d *Document) persist(ctx context.Context) error {
	if d == nil || len(d.Revisions) == 0 {
		return statusError{status: http.StatusBadRequest, error: errors.New("document has no revisions")}
	}
	docID := EscapeID(d.ID)
	for _, rev := range d.Revisions {
		if rev.path != "" {
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := rev.persist(ctx, filepath.Join(d.cdb.root, "."+docID, rev.Rev.String())); err != nil {
			return err
		}
	}

	// Make sure the winner is in the first position
	sort.Sort(d.Revisions)

	winningRev := d.Revisions[0]
	winningPath := filepath.Join(d.cdb.root, docID)
	if winningPath+filepath.Ext(winningRev.path) == winningRev.path {
		// Winner already in place, our job is done here
		return nil
	}

	// See if some other rev is currently the winning rev, and move it if necessary
	for _, rev := range d.Revisions[1:] {
		if winningPath+filepath.Ext(rev.path) == rev.path {
			if err := ctx.Err(); err != nil {
				return err
			}
			// We need to move this rev
			revpath := filepath.Join(d.cdb.root, "."+EscapeID(d.ID), rev.Rev.String())
			if err := d.cdb.fs.Mkdir(revpath, tempPerms); err != nil && !os.IsExist(err) {
				return err
			}
			// First move attachments, since they can exit both places legally.
			for attname, att := range rev.Attachments {
				if !strings.HasPrefix(att.path, rev.path+"/") {
					// This attachment is part of another rev, so skip it
					continue
				}
				newpath := filepath.Join(revpath, attname)
				if err := d.cdb.fs.Rename(att.path, newpath); err != nil {
					return err
				}
				att.path = newpath
			}
			// Try to remove the attachments dir, but don't worry if we fail.
			_ = d.cdb.fs.Remove(rev.path + "/")
			// Then make the move final by moving the json doc
			if err := d.cdb.fs.Rename(rev.path, revpath+filepath.Ext(rev.path)); err != nil {
				return err
			}
			// And remove the old rev path, if it's empty
			_ = d.cdb.fs.Remove(filepath.Dir(rev.path))
			break
		}
	}

	// Now finally put the new winner in place, first the doc, then attachments
	if err := d.cdb.fs.Rename(winningRev.path, winningPath+filepath.Ext(winningRev.path)); err != nil {
		return err
	}
	winningRev.path = winningPath + filepath.Ext(winningRev.path)

	if err := d.cdb.fs.Mkdir(winningPath, tempPerms); err != nil && !os.IsExist(err) {
		return err
	}
	revpath := filepath.Join(d.cdb.root, "."+EscapeID(d.ID), winningRev.Rev.String()) + "/"
	for attname, att := range winningRev.Attachments {
		if !strings.HasPrefix(att.path, revpath) {
			// This attachment is part of another rev, so skip it
			continue
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		newpath := filepath.Join(winningPath, attname)
		if err := d.cdb.fs.Rename(att.path, newpath); err != nil {
			return err
		}
		att.path = newpath
	}
	// And remove the old rev path, if it's empty
	_ = d.cdb.fs.Remove(filepath.Dir(revpath))
	_ = d.cdb.fs.Remove(filepath.Dir(filepath.Dir(revpath)))

	return nil
}

// leaves returns a map of leave revid to rev
func (d *Document) leaves() map[string]*Revision {
	if len(d.Revisions) == 1 {
		return map[string]*Revision{
			d.Revisions[0].Rev.String(): d.Revisions[0],
		}
	}
	leaves := make(map[string]*Revision, len(d.Revisions))
	for _, rev := range d.Revisions {
		leaves[rev.Rev.String()] = rev
	}
	for _, rev := range d.Revisions {
		// Skp over the known leaf
		for _, revid := range rev.RevHistory.Ancestors()[1:] {
			delete(leaves, revid)
		}
	}
	return leaves
}
