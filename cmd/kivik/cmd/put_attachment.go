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

	"github.com/go-kivik/kivik/v4"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/input"
)

type putAttachment struct {
	contentType string
	*root
	*input.Input
}

func putAttCmd(p *put) *cobra.Command {
	c := &putAttachment{
		root:  p.root,
		Input: p.Input,
	}
	cmd := &cobra.Command{
		Use:     "attachment [dsn]/[database]/[document]/[filename]",
		Aliases: []string{"att", "attach"},
		Short:   "Put an attachment",
		Long:    `Create or update the named attachment`,
		RunE:    c.RunE,
	}

	f := cmd.Flags()
	f.StringVarP(&c.contentType, "content-type", "T", "", "Content-Type type of the attachment")

	return cmd
}

func (c *putAttachment) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	content, err := c.RawData()
	if err != nil {
		return err
	}
	db, docID, filename, err := c.conf.DBDocFilename()
	if err != nil {
		return err
	}
	c.log.Debugf("[put] Will put attachment: %s/%s/%s/%s", client.DSN(), db, docID, filename)
	return c.retry(func() error {
		att := &kivik.Attachment{
			Filename:    filename,
			ContentType: c.contentType,
			Content:     content,
		}
		rev, err := client.DB(db).PutAttachment(cmd.Context(), docID, att, c.opts())
		if err != nil {
			return err
		}
		c.log.Info(rev)
		return nil
	})
}
