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
	"encoding/json"
	"fmt"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/fsdb/v4/filesystem"
	"github.com/go-kivik/kivik/v4/driver"
)

func TestCreateDB(t *testing.T) {
	type tt struct {
		driver *fsDriver
		path   string
		status int
		err    string
		want   *client
	}
	tests := testy.NewTable()
	tests.Add("success", tt{
		driver: &fsDriver{},
		path:   "testdata",
		want: &client{
			version: &driver.Version{
				Version:     Version,
				Vendor:      Vendor,
				RawResponse: json.RawMessage(fmt.Sprintf(`{"version":"%s","vendor":{"name":"%s"}}`, Version, Vendor)),
			},
			root: "testdata",
			fs:   filesystem.Default(),
		},
	})

	tests.Run(t, func(t *testing.T, tt tt) {
		client, err := tt.driver.NewClient(tt.path, nil)
		testy.StatusError(t, tt.err, tt.status, err)
		if d := testy.DiffInterface(tt.want, client); d != nil {
			t.Error(d)
		}
	})
}
