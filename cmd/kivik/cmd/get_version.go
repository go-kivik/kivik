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
	"bytes"

	"github.com/spf13/cobra"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
)

type getVersion struct {
	*root
}

func getVersionCmd(r *root) *cobra.Command {
	c := &getVersion{
		root: r,
	}
	return &cobra.Command{
		Use:     "version [dsn]",
		Aliases: []string{"ver"},
		Short:   "Print server version information",
		Long:    "Print server version for the provided context",
		RunE:    c.RunE,
	}
}

func (c *getVersion) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	c.conf.Finalize()

	return c.retry(func() error {
		ver, err := client.Version(cmd.Context())
		if err != nil {
			return err
		}

		format := `Server Version {{ .Version }}, {{ .Vendor }}`
		result := output.TemplateReader(format, ver, bytes.NewReader(ver.RawResponse))
		return c.fmt.Output(result)
	})
}
