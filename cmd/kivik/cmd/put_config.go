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
	"io"
	"strings"

	"github.com/spf13/cobra"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/input"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
)

type putConfig struct {
	*root
	*input.Input
	node, key string
}

func putConfigCmd(p *put) *cobra.Command {
	c := &putConfig{
		root:  p.root,
		Input: p.Input,
	}
	cmd := &cobra.Command{
		Use:   "config [dsn]",
		Short: "Set server config",
		Long:  "Sets server config at the section/key provided by the --key flag to the raw value provided by --data/--data-file",
		RunE:  c.RunE,
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&c.node, "node", "n", "_local", "Specify the node name to query")
	pf.StringVarP(&c.key, "key", "k", "", "Set the value at the specified config section and key, slash-separated")

	return cmd
}

func (c *putConfig) RunE(cmd *cobra.Command, _ []string) error {
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

	parts := strings.SplitN(c.key, "/", 2) //nolint:gomnd
	if len(parts) != 2 {                   //nolint:gomnd
		return errors.Code(errors.ErrUsage, "section/key must contain a slash")
	}

	valueReader, err := c.RawData()
	if err != nil {
		return err
	}
	value, err := io.ReadAll(valueReader)
	if err != nil {
		return err
	}

	var oldValue string
	return c.retry(func() error {
		oldValue, err = client.SetConfigValue(cmd.Context(), c.node, parts[0], parts[1], string(value))
		if err != nil {
			return err
		}

		result := output.JSONReader(oldValue)
		return c.fmt.Output(result)
	})
}
