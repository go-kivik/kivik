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
	"net/url"
	"strings"

	"github.com/spf13/cobra"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
)

type getConfig struct {
	*root
	node, key string
}

func getConfigCmd(r *root) *cobra.Command {
	c := &getConfig{
		root: r,
	}
	cmd := &cobra.Command{
		Use:   "config [dsn]",
		Short: "Get server config",
		RunE:  c.RunE,
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&c.node, "node", "n", "_local", "Specify the node name to query")
	pf.StringVarP(&c.key, "key", "k", "", "Fetch only the specified config section, and optionally key, separated by a slash")

	return cmd
}

func configFromDSN(dsn *url.URL) (node, key string, ok bool) {
	parts := strings.Split(dsn.Path, "/")
	if len(parts) < 4 || parts[1] != "_node" || parts[3] != "_config" {
		return "", "", false
	}
	return parts[2], strings.Join(parts[4:], "/"), true
}

func (c *getConfig) RunE(cmd *cobra.Command, _ []string) error {
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

	return c.retry(func() error {
		var conf interface{}
		var err error
		if c.key != "" {
			if parts := strings.SplitN(c.key, "/", 2); len(parts) > 1 { //nolint:gomnd
				conf, err = client.ConfigValue(cmd.Context(), c.node, parts[0], parts[1])
			} else {
				conf, err = client.ConfigSection(cmd.Context(), c.node, c.key)
			}
		} else {
			conf, err = client.Config(cmd.Context(), c.node)
		}
		if err != nil {
			return err
		}

		result := output.JSONReader(conf)
		return c.fmt.Output(result)
	})
}
