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
)

type postCompact struct {
	*root
}

func postCompactCmd(r *root) *cobra.Command {
	c := &postCompact{
		root: r,
	}
	cmd := &cobra.Command{
		Use:   "compact [dsn]/[database]",
		Short: "Compact the database",
		Long:  `Compact the disk database file by pruning unused data`,
		RunE:  c.RunE,
	}

	return cmd
}

func (c *postCompact) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	dsn, err := c.conf.URL()
	if err != nil {
		return err
	}
	_, db := dbCommandFromDSN(dsn)
	if db == "" {
		db, err = c.conf.DB()
		if err != nil {
			return err
		}
	}
	c.log.Debugf("[post] Will compact: %s/%s", client.DSN(), db)
	return c.retry(func() error {
		err := client.DB(db).Compact(cmd.Context())
		if err != nil {
			return err
		}
		return c.fmt.OK()
	})
}
