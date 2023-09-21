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
	"github.com/spf13/cobra"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
)

type getAllDBs struct {
	*root
}

func getAllDBsCmd(r *root) *cobra.Command {
	c := &getAllDBs{
		root: r,
	}
	return &cobra.Command{
		Use:     "all-dbs [dsn]",
		Aliases: []string{"alldbs"},
		Short:   "List all databases",
		RunE:    c.RunE,
	}
}

func (c *getAllDBs) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	c.conf.Finalize()

	return c.retry(func() error {
		dbs, err := client.AllDBs(cmd.Context(), c.opts())
		if err != nil {
			return err
		}

		format := `{{ range . }}{{ . }}
{{ end }}`
		result := output.TemplateReader(format, dbs, output.JSONReader(dbs))
		return c.fmt.Output(result)
	})
}
