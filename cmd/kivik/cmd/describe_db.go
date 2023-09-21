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
)

type descrDB struct {
	*root
}

func descrDBCmd(r *root) *cobra.Command {
	g := &descrDB{
		root: r,
	}
	return &cobra.Command{
		Use:     "database [dsn]/[database]",
		Aliases: []string{"db"},
		Short:   "Describe a database",
		Long:    `Fetch information about a database`,
		RunE:    g.RunE,
	}
}

func (c *descrDB) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	db, _, err := c.conf.DBDoc()
	if err != nil {
		return err
	}
	c.log.Debugf("[get] Will fetch database: %s/%s", client.DSN(), db)
	return c.retry(func() error {
		stats, err := client.DB(db).Stats(cmd.Context())
		if err != nil {
			return err
		}
		return c.fmt.Output(bytes.NewReader(stats.RawResponse))
	})
}
