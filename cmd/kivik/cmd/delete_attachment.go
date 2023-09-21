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

	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
)

type deleteAtt struct {
	*root
}

func deleteAttachmentCmd(r *root) *cobra.Command {
	c := &deleteAtt{
		root: r,
	}
	return &cobra.Command{
		Use:     "attachment [dsn]/[database]/[document]/[filename]",
		Aliases: []string{"att", "attach"},
		Short:   "Delete an attachment",
		RunE:    c.RunE,
	}
}

func (c *deleteAtt) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	db, docID, filename, err := c.conf.DBDocFilename()
	if err != nil {
		return err
	}
	c.log.Debugf("[delete] Will delete document: %s/%s/%s", client.DSN(), db, docID)
	return c.retry(func() error {
		newRev, err := client.DB(db).DeleteAttachment(cmd.Context(), docID, "", filename, c.opts())
		if err != nil {
			return err
		}

		return c.fmt.Output(output.JSONReader(map[string]interface{}{
			"ok":  true,
			"id":  docID,
			"rev": newRev,
		}))
	})
}
