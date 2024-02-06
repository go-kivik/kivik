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
			got, err := calculateRev(tt.docID, tt.doc)
			if !testy.ErrorMatches(tt.wantErr, err) {
				t.Errorf("unexpected error = %v, wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("unexpected rev= %v, want %v", got, tt.want)
			}
		})
	}
}
