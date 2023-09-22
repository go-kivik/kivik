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
)

type postCompactViews struct {
	*root
}

func postCompactViewsCmd(r *root) *cobra.Command {
	c := &postCompactViews{
		root: r,
	}
	cmd := &cobra.Command{
		Use:     "compact-views [dsn]/[database]/[design-doc]",
		Aliases: []string{"compact-view", "cv"},
		Short:   "Compact the database",
		Long:    `Compact the disk database file by pruning unused data`,
		RunE:    c.RunE,
	}

	return cmd
}

func compactViewFromDSN(dsn *url.URL) (db, ddoc string) {
	parts := strings.Split(dsn.Path, "/")
	if len(parts) != 4 || parts[2] != "_compact" {
		return "", ""
	}
	return parts[1], parts[3]
}

func (c *postCompactViews) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	dsn, err := c.conf.URL()
	if err != nil {
		return err
	}
	db, ddoc := compactViewFromDSN(dsn)
	if db == "" {
		db, ddoc, err = c.conf.DBDoc()
		if err != nil {
			return err
		}
		ddoc = strings.TrimPrefix(ddoc, "_design/")
	}
	c.log.Debugf("[post] Will compact: %s/%s", client.DSN(), db)
	return c.retry(func() error {
		err := client.DB(db).CompactView(cmd.Context(), ddoc)
		if err != nil {
			return err
		}
		return c.fmt.OK()
	})
}
