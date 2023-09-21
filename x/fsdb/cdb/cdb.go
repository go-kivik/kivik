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

// Package cdb provides the core CouchDB types.
package cdb

import (
	"fmt"
)

// RevHistory represents the recent ancestors of a revision.
type RevHistory struct {
	Start int64    `json:"start" yaml:"start"`
	IDs   []string `json:"ids" yaml:"ids"`
}

// Ancestors returns the full, known revision history, newest first, starting
// with the current rev.
func (h *RevHistory) Ancestors() []string {
	history := make([]string, len(h.IDs))
	for i, id := range h.IDs {
		history[i] = fmt.Sprintf("%d-%s", h.Start-int64(i), id)
	}
	return history
}

// AddRevision returns a new history, with rev added.
func (h *RevHistory) AddRevision(rev RevID) *RevHistory {
	return &RevHistory{
		Start: rev.Seq,
		IDs:   append([]string{rev.Sum}, h.IDs...),
	}
}
