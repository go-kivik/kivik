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

type postViewCleanup struct {
	*root
}

func postViewCleanupCmd(r *root) *cobra.Command {
	c := &postViewCleanup{
		root: r,
	}
	cmd := &cobra.Command{
		Use:     "view-cleanup [dsn]/[database]",
		Aliases: []string{"vc"},
		Short:   "Removes unused view index files",
		Long:    `Removes view index files that are no longer required by CouchDB as a result of changed views within design documents.`,
		RunE:    c.RunE,
	}

	return cmd
}

func (c *postViewCleanup) RunE(cmd *cobra.Command, _ []string) error {
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
	c.log.Debugf("[post] Will perform view cleanup for: %s/%s", client.DSN(), db)
	return c.retry(func() error {
		err := client.DB(db).ViewCleanup(cmd.Context())
		if err != nil {
			return err
		}
		return c.fmt.OK()
	})
}
