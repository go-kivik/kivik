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
	"net/http"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/x/fsdb/filesystem"
)

func TestAttachmentsIterator(t *testing.T) {
	type tt struct {
		r      *Revision
		status int
		err    string
	}
	tests := testy.NewTable()
	tests.Add("missing attachment", tt{
		r: &Revision{
			options: map[string]interface{}{
				"attachments": true,
			},
			RevMeta: RevMeta{
				Attachments: map[string]*Attachment{
					"notfound.txt": {
						fs:   filesystem.Default(),
						path: "/somewhere/notfound.txt",
					},
				},
			},
		},
		status: http.StatusInternalServerError,
		err:    "open /somewhere/notfound.txt: no such file or directory",
	})
	tests.Run(t, func(t *testing.T, tt tt) {
		_, err := tt.r.AttachmentsIterator()
		testy.StatusError(t, tt.err, tt.status, err)
	})
}
