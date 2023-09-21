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

type postFlush struct {
	*root
}

func postFlushCmd(r *root) *cobra.Command {
	c := &postFlush{
		root: r,
	}
	cmd := &cobra.Command{
		Use:     "flush [dsn]/[database]",
		Aliases: []string{"ensure-full-commit"},
		Short:   "Commit recent changes",
		Long:    `Commit changes to the database in case the delayed_commits=true option was set. For CouchDB 3.0+ delayed_commits is always false, and this operation is a no-op.`,
		RunE:    c.RunE,
	}

	return cmd
}

func (c *postFlush) RunE(cmd *cobra.Command, _ []string) error {
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
	c.log.Debugf("[post] Will flush for: %s/%s", client.DSN(), db)
	return c.retry(func() error {
		err := client.DB(db).Flush(cmd.Context())
		if err != nil {
			return err
		}
		return c.fmt.OK()
	})
}
