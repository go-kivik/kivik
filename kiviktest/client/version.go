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

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kiviktest/v4/kt"
)

func init() {
	kt.Register("Version", version)
}

func version(ctx *kt.Context) {
	ctx.RunAdmin(func(ctx *kt.Context) {
		testServerInfo(ctx, ctx.Admin)
	})
	ctx.RunNoAuth(func(ctx *kt.Context) {
		testServerInfo(ctx, ctx.NoAuth)
	})
}

func testServerInfo(ctx *kt.Context, client *kivik.Client) {
	ctx.Parallel()
	info, err := client.Version(context.Background())
	if !ctx.IsExpectedSuccess(err) {
		return
	}
	version := regexp.MustCompile(ctx.MustString("version"))
	vendor := regexp.MustCompile(ctx.MustString("vendor"))
	if !version.MatchString(info.Version) {
		ctx.Errorf("Version '%s' does not match /%s/", info.Version, version)
	}
	if !vendor.MatchString(info.Vendor) {
		ctx.Errorf("Vendor '%s' doesnot match /%s/", info.Vendor, vendor)
	}
}
