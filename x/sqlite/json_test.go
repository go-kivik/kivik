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
	"testing"

	"gitlab.com/flimzy/testy"
)

func Test_calculateRev(t *testing.T) {
	tests := []struct {
		name    string
		docID   string
		doc     interface{}
		want    string
		wantErr string
	}{
		{
			name: "no rev in document",
			doc:  map[string]string{"foo": "bar"},
			want: "9bb58f26192e4ba00f01e2e7b136bbd8",
		},
		{
			name: "rev in document",
			doc: map[string]interface{}{
				"_rev": "1-1234567890abcdef1234567890abcdef",
				"foo":  "bar",
			},
			want: "9bb58f26192e4ba00f01e2e7b136bbd8",
		},
		{
			name:  "add docID",
			docID: "foo",
			doc:   map[string]string{"foo": "bar"},
			want:  "6fe51f74859f3579abaccc426dd5104f",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _, err := prepareDoc(tt.docID, tt.doc)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("unexpected rev= %v, want %v", got, tt.want)
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
			wantErr: "missing _rev",
		},
		{
			name:    "empty",
			doc:     map[string]string{},
			wantErr: "missing _rev",
		},
		{
			name:    "no rev",
			doc:     map[string]string{"foo": "bar"},
			wantErr: "missing _rev",
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
			name:    "invalid rev",
			doc:     map[string]string{"_rev": "foo"},
			wantErr: "strconv.ParseInt: parsing \"foo\": invalid syntax",
		},
		{
			name:    "rev id only",
			doc:     map[string]string{"_rev": "1"},
			wantRev: "1-",
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
			if rev.String() != tt.wantRev {
				t.Errorf("unexpected rev= %v, want %v", rev, tt.wantRev)
			}
		})
	}
}
