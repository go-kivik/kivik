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

package client

import (
	"context"
	"regexp"
	"testing"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	kt.Register("Version", version)
}

func version(t *testing.T, c *kt.Context) {
	t.Helper()
	c.RunAdmin(t, func(t *testing.T) {
		t.Helper()
		testServerInfo(t, c, c.Admin)
	})
	c.RunNoAuth(t, func(t *testing.T) {
		t.Helper()
		testServerInfo(t, c, c.NoAuth)
	})
}

func testServerInfo(t *testing.T, c *kt.Context, client *kivik.Client) { //nolint:thelper
	t.Parallel()
	info, err := client.Version(context.Background())
	if !c.IsExpectedSuccess(t, err) {
		return
	}
	version := regexp.MustCompile(c.MustString(t, "version"))
	vendor := regexp.MustCompile(c.MustString(t, "vendor"))
	if !version.MatchString(info.Version) {
		t.Errorf("Version '%s' does not match /%s/", info.Version, version)
	}
	if !vendor.MatchString(info.Vendor) {
		t.Errorf("Vendor '%s' does not match /%s/", info.Vendor, vendor)
	}
}
