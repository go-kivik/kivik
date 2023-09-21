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

package cmd

import (
	"regexp"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

func Test_replicate_RunE(t *testing.T) {
	tests := testy.NewTable()

	tests.Add("missing source", cmdTest{
		args:   []string{"replicate"},
		status: errors.ErrUsage,
	})
	tests.Add("missing target", cmdTest{
		args:   []string{"replicate", "-O", "source=./foo"},
		status: errors.ErrUsage,
	})
	tests.Add("fs to fs", func(t *testing.T) interface{} {
		var tmpdir string
		tests.Cleanup(testy.TempDir(t, &tmpdir))

		return cmdTest{
			args: []string{"replicate", "-O", "source=./testdata/source", "-O", "target=" + tmpdir},
		}
	})

	tests.Run(t, func(t *testing.T, tt cmdTest) {
		re := testy.Replacement{
			Regexp:      regexp.MustCompile(`_time": ".*?"`),
			Replacement: `_time": "xxx"`,
		}
		tt.Test(t, re)
	})
}
