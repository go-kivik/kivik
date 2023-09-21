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

	"github.com/go-kivik/xkivik/v4/cmd/kivik/input"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
)

type postPurge struct {
	*root
	*input.Input
	revs []string
}

func postPurgeRootCmd(r *root) *cobra.Command {
	c := &postPurge{
		root:  r,
		Input: input.New(),
	}
	cmd := &cobra.Command{
		Use:     "purge [dsn]/[database]/[document]",
		Aliases: []string{"ensure-full-commit"},
		Short:   "Purge document revision(s)",
		Long:    `Permanently remove the references to documents in the database. Provide the document ID in the DSN, or pass a map of document IDs to revisions via --data or similar.`,
		RunE:    c.RunE,
	}

	c.Input.ConfigFlags(cmd.PersistentFlags())

	pf := cmd.PersistentFlags()
	pf.StringSliceVarP(&c.revs, "revs", "R", nil, "List of revisions to purge")

	return cmd
}

func postPurgeCmd(p *post) *cobra.Command {
	c := &postPurge{
		root:  p.root,
		Input: p.Input,
	}
	cmd := &cobra.Command{
		Use:     "purge [dsn]/[database]/[document]",
		Aliases: []string{"ensure-full-commit"},
		Short:   "Purge document revision(s)",
		Long:    `Permanently remove the references to documents in the database. Provide the document ID in the DSN, or pass a map of document IDs to revisions via --data or similar.`,
		RunE:    c.RunE,
	}

	pf := cmd.PersistentFlags()
	pf.StringSliceVarP(&c.revs, "revs", "R", nil, "List of revisions to purge")

	return cmd
}

func (c *postPurge) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	var docRevMap map[string][]string
	var db string
	if c.HasInput() {
		err := c.As(&docRevMap)
		if err != nil {
			return err
		}
		dsn, err := c.conf.URL()
		if err != nil {
			return err
		}
		if cmd, dsnDB := dbCommandFromDSN(dsn); cmd == "_purge" {
			db = dsnDB
		}
		if db == "" {
			db, err = c.conf.DB()
			if err != nil {
				return err
			}
		}
	} else {
		var doc string
		db, doc, err = c.conf.DBDoc()
		if err != nil {
			return err
		}
		docRevMap = map[string][]string{
			doc: c.revs,
		}
	}
	c.log.Debugf("[post] Will purge: %s/%s (%v)", client.DSN(), db, docRevMap)
	return c.retry(func() error {
		result, err := client.DB(db).Purge(cmd.Context(), docRevMap)
		if err != nil {
			return err
		}
		return c.fmt.Output(output.JSONReader(result))
	})
}
