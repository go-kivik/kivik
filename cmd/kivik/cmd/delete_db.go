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

type deleteDB struct {
	*root
}

func deleteDBCmd(r *root) *cobra.Command {
	c := &deleteDB{
		root: r,
	}
	return &cobra.Command{
		Use:     "database [dsn]/[database]",
		Aliases: []string{"db"},
		Short:   "Delete a database",
		RunE:    c.RunE,
	}
}

func (c *deleteDB) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	db, _, err := c.conf.DBDoc()
	if err != nil {
		return err
	}
	c.log.Debugf("[delete] Will delete database: %s/%s/%s", client.DSN(), db)
	return c.retry(func() error {
		err := client.DestroyDB(cmd.Context(), db, c.opts())
		if err != nil {
			return err
		}

		return c.fmt.OK()
	})
}
