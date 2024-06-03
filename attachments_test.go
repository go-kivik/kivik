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

package kivik

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
	"github.com/go-kivik/kivik/v4/int/mock"
)

func TestAttachmentMarshalJSON(t *testing.T) {
	type tst struct {
		att      *Attachment
		expected string
		err      string
	}
	tests := testy.NewTable()
	tests.Add("foo.txt", tst{
		att: &Attachment{
			Content:     io.NopCloser(strings.NewReader("test attachment\n")),
			Filename:    "foo.txt",
			ContentType: "text/plain",
		},
		expected: `{
			"content_type": "text/plain",
			"data": "dGVzdCBhdHRhY2htZW50Cg=="
		}`,
	})
	tests.Add("revpos", tst{
		att: &Attachment{
			Content:     io.NopCloser(strings.NewReader("test attachment\n")),
			Filename:    "foo.txt",
			ContentType: "text/plain",
			RevPos:      3,
		},
		expected: `{
			"content_type": "text/plain",
			"data": "dGVzdCBhdHRhY2htZW50Cg==",
			"revpos": 3
		}`,
	})
	tests.Add("follows", tst{
		att: &Attachment{
			Content:     io.NopCloser(strings.NewReader("test attachment\n")),
			Filename:    "foo.txt",
			Follows:     true,
			ContentType: "text/plain",
			RevPos:      3,
		},
		expected: `{
			"content_type": "text/plain",
			"follows": true,
			"revpos": 3
		}`,
	})
	tests.Add("read error", tst{
		att: &Attachment{
			Content:     io.NopCloser(&errorReader{}),
			Filename:    "foo.txt",
			ContentType: "text/plain",
		},
		err: "json: error calling MarshalJSON for type *kivik.Attachment: errorReader",
	})
	tests.Add("stub", tst{
		att: &Attachment{
			Content:     io.NopCloser(strings.NewReader("content")),
			Stub:        true,
			Filename:    "foo.txt",
			ContentType: "text/plain",
			Size:        7,
		},
		expected: `{
			"content_type": "text/plain",
			"length": 7,
			"stub": true
		}`,
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := json.Marshal(test.att)
		if !testy.ErrorMatches(test.err, err) {
			t.Errorf("Unexpected error: %s", err)
		}
		if d := testy.DiffJSON([]byte(test.expected), result); d != nil {
			t.Error(d)
		}
	})
}

func TestAttachmentUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string

		body     string
		expected *Attachment
		err      string
	}{
		{
			name: "stub",
			input: `{
					"content_type": "text/plain",
					"stub": true
				}`,
			expected: &Attachment{
				ContentType: "text/plain",
				Stub:        true,
			},
		},
		{
			name: "simple",
			input: `{
					"content_type": "text/plain",
					"data": "dGVzdCBhdHRhY2htZW50Cg=="
				}`,
			body: "test attachment\n",
			expected: &Attachment{
				ContentType: "text/plain",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := new(Attachment)
			err := json.Unmarshal([]byte(test.input), result)
			if !testy.ErrorMatches(test.err, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			var body []byte
			content := result.Content
			t.Cleanup(func() {
				_ = content.Close()
			})
			body, err = io.ReadAll(result.Content)
			if err != nil {
				t.Fatal(err)
			}
			result.Content = nil
			if d := testy.DiffText(test.body, string(body)); d != nil {
				t.Errorf("Unexpected body:\n%s", d)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Errorf("Unexpected result:\n%s", d)
			}
		})
	}
}

func TestAttachmentsUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name  string
		input string

		expected Attachments
		err      string
	}{
		{
			name:     "no attachments",
			input:    "{}",
			expected: Attachments{},
		},
		{
			name: "one attachment",
			input: `{
				"foo.txt": {
					"content_type": "text/plain",
					"data": "dGVzdCBhdHRhY2htZW50Cg=="
				}
			}`,
			expected: Attachments{
				"foo.txt": &Attachment{
					Filename:    "foo.txt",
					ContentType: "text/plain",
				},
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var att Attachments
			err := json.Unmarshal([]byte(test.input), &att)
			if !testy.ErrorMatches(test.err, err) {
				t.Errorf("Unexpected error: %s", err)
			}
			for _, v := range att {
				_ = v.Content.Close()
				v.Content = nil
			}
			if d := testy.DiffInterface(test.expected, att); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestAttachmentsIteratorNext(t *testing.T) {
	tests := []struct {
		name     string
		iter     *AttachmentsIterator
		expected *Attachment
		status   int
		err      string
	}{
		{
			name: "error",
			iter: &AttachmentsIterator{
				atti: &mock.Attachments{
					NextFunc: func(_ *driver.Attachment) error {
						return &internal.Error{Status: http.StatusBadGateway, Err: errors.New("error")}
					},
				},
			},
			status: http.StatusBadGateway,
			err:    "error",
		},
		{
			name: "success",
			iter: &AttachmentsIterator{
				atti: &mock.Attachments{
					NextFunc: func(att *driver.Attachment) error {
						*att = driver.Attachment{
							Filename: "foo.txt",
						}
						return nil
					},
				},
			},
			expected: &Attachment{
				Filename: "foo.txt",
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.iter.Next()
			if d := internal.StatusErrorDiff(test.err, test.status, err); d != "" {
				t.Error(d)
			}
			if d := testy.DiffInterface(test.expected, result); d != nil {
				t.Error(d)
			}
		})
	}
}

func TestAttachments_Next_resets_iterator_value(t *testing.T) {
	idx := 0
	atts := &AttachmentsIterator{
		atti: &mock.Attachments{
			NextFunc: func(att *driver.Attachment) error {
				idx++
				switch idx {
				case 1:
					att.Filename = strconv.Itoa(idx)
					return nil
				case 2:
					return nil
				}
				return io.EOF
			},
		},
	}

	wantFilenames := []string{"1", ""}
	gotFilenames := []string{}
	for {
		att, err := atts.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}
		gotFilenames = append(gotFilenames, att.Filename)

	}
	if d := cmp.Diff(wantFilenames, gotFilenames); d != "" {
		t.Error(d)
	}
}
