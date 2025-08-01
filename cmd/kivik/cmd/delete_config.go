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
	"strings"

	"github.com/spf13/cobra"

	"github.com/go-kivik/kivik/v4/cmd/kivik/errors"
	"github.com/go-kivik/kivik/v4/cmd/kivik/output"
)

type deleteConfig struct {
	*root
	node, key string
}

func deleteConfigCmd(r *root) *cobra.Command {
	c := &deleteConfig{
		root: r,
	}
	cmd := &cobra.Command{
		Use:   "config [dsn]",
		Short: "Delete server config",
		RunE:  c.RunE,
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&c.node, "node", "n", "_local", "Specify the node name to query")
	pf.StringVarP(&c.key, "key", "k", "", "Delete the value stored at the specified config section and key, slash-separated")

	return cmd
}

func (c *deleteConfig) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	dsn, err := c.conf.URL()
	if err != nil {
		return err
	}
	c.conf.Finalize()
	if node, key, ok := configFromDSN(dsn); ok {
		c.node = node
		c.key = key
	}

	parts := strings.SplitN(c.key, "/", 2)
	if len(parts) != 2 {
		return errors.Code(errors.ErrUsage, "section/key must contain a slash")
	}

	var oldValue string
	return c.retry(func() error {
		oldValue, err = client.DeleteConfigKey(cmd.Context(), c.node, parts[0], parts[1])
		if err != nil {
			return err
		}

		result := output.JSONReader(oldValue)
		return c.fmt.Output(result)
	})
}
