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
	"runtime"

	"github.com/spf13/cobra"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
	v "github.com/go-kivik/xkivik/v4/cmd/kivik/version"
)

type version struct {
	*root
}

func versionCmd(r *root) *cobra.Command {
	c := &version{
		root: r,
	}
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"ver"},
		Short:   "Print client and server version information",
		Long:    "Print client and server versions for the provided context",
		RunE:    c.RunE,
	}
}

func (c *version) RunE(cmd *cobra.Command, _ []string) error {
	c.conf.Finalize()

	data := struct {
		Version   string `json:"version"`
		GoVersion string `json:"goVersion"`
		GOARCH    string `json:"GOARCH"`
		GOOS      string `json:"GOOS"`
	}{
		Version:   v.Version,
		GoVersion: runtime.Version(),
		GOOS:      runtime.GOOS,
		GOARCH:    runtime.GOARCH,
	}

	format := `kivik version {{ .Version }} {{ .GoVersion }} {{ .GOOS }}/{{ .GOARCH }}`
	result := output.TemplateReader(format, data, output.JSONReader(data))
	return c.fmt.Output(result)
}
