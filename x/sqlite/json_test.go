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

//go:build !js
// +build !js

package sqlite

import (
	"io"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"
)

func mustParseMD5sum(s string) md5sum { //nolint:unparam
	m, err := parseMD5sum(s)
	if err != nil {
		panic(err)
	}
	return m
}

func Test_prepareDoc(t *testing.T) {
	tests := []struct {
		name    string
		docID   string
		doc     interface{}
		want    *docData
		wantErr string
	}{
		{
			name: "no rev in document",
			doc:  map[string]string{"foo": "bar"},
			want: &docData{
				Doc:    []byte(`{"foo":"bar"}`),
				MD5sum: mustParseMD5sum("9bb58f26192e4ba00f01e2e7b136bbd8"),
			},
		},
		{
			name: "rev in document",
			doc: map[string]interface{}{
				"_rev": "1-1234567890abcdef1234567890abcdef",
				"foo":  "bar",
			},
			want: &docData{
				Doc:    []byte(`{"foo":"bar"}`),
				MD5sum: mustParseMD5sum("9bb58f26192e4ba00f01e2e7b136bbd8"),
			},
		},
		{
			name:  "add docID",
			docID: "foo",
			doc:   map[string]string{"foo": "bar"},
			want: &docData{
				ID:     "foo",
				Doc:    []byte(`{"foo":"bar"}`),
				MD5sum: mustParseMD5sum("9bb58f26192e4ba00f01e2e7b136bbd8"),
			},
		},
		{
			name: "deleted true",
			doc: map[string]interface{}{
				"_rev":     "1-1234567890abcdef1234567890abcdef",
				"_deleted": true,
				"foo":      "bar",
			},
			want: &docData{
				Doc:     []byte(`{"foo":"bar"}`),
				Deleted: true,
				MD5sum:  mustParseMD5sum("9bb58f26192e4ba00f01e2e7b136bbd8"),
			},
		},
		{
			name: "wrong type for _deleted",
			doc: map[string]interface{}{
				"_rev":     "1-1234567890abcdef1234567890abcdef",
				"_deleted": "oink",
				"foo":      "bar",
			},
			wantErr: "json: cannot unmarshal string into Go struct field docData._deleted of type bool",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := prepareDoc(tt.docID, tt.doc)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf(d)
			}
		})
	}
}

func Test_extractRev(t *testing.T) {
	tests := []struct {
		name    string
		doc     interface{}
		wantRev string
		wantErr string
	}{
		{
			name:    "nil",
			doc:     nil,
			wantRev: "",
		},
		{
			name:    "empty",
			doc:     map[string]string{},
			wantRev: "",
		},
		{
			name:    "no rev",
			doc:     map[string]string{"foo": "bar"},
			wantRev: "",
		},
		{
			name:    "rev in string",
			doc:     map[string]string{"_rev": "1-1234567890abcdef1234567890abcdef"},
			wantRev: "1-1234567890abcdef1234567890abcdef",
		},
		{
			name:    "rev in interface",
			doc:     map[string]interface{}{"_rev": "1-1234567890abcdef1234567890abcdef"},
			wantRev: "1-1234567890abcdef1234567890abcdef",
		},
		{
			name: "rev in struct",
			doc: struct {
				Rev string `json:"_rev"`
			}{Rev: "1-1234567890abcdef1234567890abcdef"},
			wantRev: "1-1234567890abcdef1234567890abcdef",
		},
		{
			name:    "rev id only",
			doc:     map[string]string{"_rev": "1"},
			wantRev: "1",
		},
		{
			name:    "invalid rev struct",
			doc:     struct{ Rev func() }{},
			wantErr: "json: unsupported type: func()",
		},
		{
			name: "invalid rev type",
			doc: struct {
				Rev int `json:"_rev"`
			}{Rev: 1},
			wantErr: "json: cannot unmarshal number into Go struct field ._rev of type string",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rev, err := extractRev(tt.doc)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil {
				return
			}
			if rev != tt.wantRev {
				t.Errorf("unexpected rev= %v, want %v", rev, tt.wantRev)
			}
		})
	}
}

func Test_revsInfo_revs(t *testing.T) {
	tests := []struct {
		name string
		ri   revsInfo
		want []string
	}{
		{
			name: "empty",
			ri:   revsInfo{},
			want: []string{},
		},
		{
			name: "single",
			ri: revsInfo{
				Start: 1,
				IDs:   []string{"a"},
			},
			want: []string{
				"1-a",
			},
		},
		{
			name: "multiple",
			ri: revsInfo{
				Start: 8,
				IDs:   []string{"z", "y", "x"},
			},
			want: []string{
				"6-x",
				"7-y",
				"8-z",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := []string{}
			for _, r := range tt.ri.revs() {
				got = append(got, r.String())
			}
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Errorf(d)
			}
		})
	}
}

func Test_mergeIntoDoc(t *testing.T) {
	tests := []struct {
		name string
		doc  fullDoc
		want string
	}{
		{
			name: "nothing to merge",
			doc:  fullDoc{Doc: []byte(`{"foo":"bar"}`)},
			want: `{"foo":"bar"}`,
		},
		{
			name: "id and rev",
			doc: fullDoc{
				ID:  "foo",
				Rev: "1-abc",
				Doc: []byte(`{"foo":"bar"}`),
			},
			want: `{"_id":"foo","_rev":"1-abc","foo":"bar"}`,
		},
		{
			name: "id, rev, and other",
			doc: fullDoc{
				ID:       "foo",
				Rev:      "1-abc",
				Doc:      []byte(`{"foo":"bar"}`),
				LocalSeq: 1,
			},
			want: `{"_id":"foo","_rev":"1-abc","foo":"bar","_local_seq":1}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := io.ReadAll(mergeIntoDoc(tt.doc))
			if d := cmp.Diff(tt.want, string(got)); d != "" {
				t.Errorf(d)
			}
		})
	}
}
