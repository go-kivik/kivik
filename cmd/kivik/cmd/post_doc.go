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
)

type postDoc struct {
	*root
	*input.Input
}

func postDocCmd(p *post) *cobra.Command {
	c := &postDoc{
		root:  p.root,
		Input: p.Input,
	}
	cmd := &cobra.Command{
		Use:     "document [dsn]/[database]",
		Aliases: []string{"doc"},
		Short:   "Create a document",
		Long:    `Create a document with sever-assigned ID`,
		RunE:    c.RunE,
	}

	return cmd
}

func (c *postDoc) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	db, err := c.conf.DB()
	if err != nil {
		return err
	}
	doc, err := c.JSONData()
	if err != nil {
		return err
	}
	c.log.Debugf("[post] Will post document to: %s/%s", client.DSN(), db)
	return c.retry(func() error {
		docID, rev, err := client.DB(db).CreateDoc(cmd.Context(), doc, c.opts())
		if err != nil {
			return err
		}
		return c.fmt.UpdateResult(docID, rev)
	})
}
