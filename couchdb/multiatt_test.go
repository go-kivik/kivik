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

package couchdb

import (
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
)

type attStruct struct {
	Attachments kivik.Attachments `json:"_attachments"`
}

type attPtrStruct struct {
	Attachments *kivik.Attachments `json:"_attachments"`
}

type wrongTypeStruct struct {
	Attachments string `json:"_attachments"`
}

type wrongTagStruct struct {
	Attachments kivik.Attachments `json:"foo"`
}

func TestExtractAttachments(t *testing.T) {
	tests := []struct {
		name string
		doc  interface{}

		expected *kivik.Attachments
		ok       bool
	}{
		{
			name:     "no attachments",
			doc:      map[string]interface{}{"foo": "bar"},
			expected: nil,
			ok:       false,
		},
		{
			name: "in map",
			doc: map[string]interface{}{"_attachments": kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
			}},
			expected: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain"},
			},
			ok: true,
		},
		{
			name:     "wrong type in map",
			doc:      map[string]interface{}{"_attachments": "oink"},
			expected: nil,
			ok:       false,
		},
		{
			name:     "non standard map, non struct",
			doc:      map[string]string{"foo": "bar"},
			expected: nil,
			ok:       false,
		},
		{
			name: "attachments in struct",
			doc: attStruct{
				Attachments: kivik.Attachments{
					"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
				},
			},
			expected: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain"},
			},
			ok: true,
		},
		{
			name: "pointer to attachments in struct",
			doc: attPtrStruct{
				Attachments: &kivik.Attachments{
					"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
				},
			},
			expected: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain"},
			},
			ok: true,
		},
		{
			name: "wrong type of struct",
			doc: wrongTypeStruct{
				Attachments: "foo",
			},
			expected: nil,
			ok:       false,
		},
		{
			name: "wrong json tag",
			doc: wrongTagStruct{
				Attachments: kivik.Attachments{
					"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
				},
			},
			expected: nil,
			ok:       false,
		},
		{
			name: "pointer to struct with attachments",
			doc: &attStruct{
				Attachments: kivik.Attachments{
					"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
				},
			},
			expected: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain"},
			},
			ok: true,
		},
		{
			name: "pointer to map with attachments",
			doc: &(map[string]interface{}{"_attachments": kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
			}}),
			expected: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain"},
			},
			ok: true,
		},
		{
			name: "pointer in map",
			doc: map[string]interface{}{"_attachments": &kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain", Content: Body("test content")},
			}},
			expected: &kivik.Attachments{
				"foo.txt": &kivik.Attachment{Filename: "foo.txt", ContentType: "text/plain"},
			},
			ok: true,
		},
		{
			name: "nil doc",
			doc:  nil,
			ok:   false,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, ok := extractAttachments(test.doc)
			if ok != test.ok {
				t.Errorf("Unexpected OK: %v", ok)
			}
			if result != nil {
				for _, att := range *result {
					att.Content = nil
				}
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}
