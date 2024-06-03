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
	"testing"

	"gitlab.com/flimzy/testy"

	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func TestNewRevision(t *testing.T) {
	type tt struct {
		i      interface{}
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("simple", tt{
		i: map[string]interface{}{
			"_rev":  "1-xxx",
			"value": "foo",
		},
	})
	tests.Add("with attachments", tt{
		i: map[string]interface{}{
			"_rev": "3-asdf",
			"_attachments": map[string]interface{}{
				"foo.txt": map[string]interface{}{
					"content_type": "text/plain",
					"data":         []byte("This is some content"),
				},
			},
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		fs := &FS{}
		rev, err := fs.NewRevision(tt.i)
		if d := internal.StatusErrorDiff(tt.err, tt.status, err); d != "" {
			t.Error(d)
		}
		rev.options = map[string]interface{}{
			"revs":          true,
			"attachments":   true,
			"header:accept": "application/json",
		}
		if d := testy.DiffAsJSON(testy.Snapshot(t), rev); d != nil {
			t.Error(d)
		}
	})
}
