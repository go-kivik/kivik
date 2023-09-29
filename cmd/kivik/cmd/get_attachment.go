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
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/cmd/kivik/output"
)

type getAttachment struct {
	*root
}

func getAttachmentCmd(r *root) *cobra.Command {
	c := &getAttachment{
		root: r,
	}
	return &cobra.Command{
		Use:     "attachment [dsn]/[database]/[document]/[filename]",
		Aliases: []string{"att", "attach"},
		Short:   "Get an attachment",
		Long:    `Fetch an attachment with the HTTP GET verb`,
		RunE:    c.RunE,
	}
}

func (c *getAttachment) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	db, docID, filename, err := c.conf.DBDocFilename()
	if err != nil {
		return err
	}
	c.log.Debugf("[get] Will fetch document: %s/%s/%s", client.DSN(), db, docID)
	var att *kivik.Attachment
	err = c.retry(func() error {
		var err error
		att, err = client.DB(db).GetAttachment(cmd.Context(), docID, filename, c.opts())
		if err != nil {
			return fmt.Errorf("client.DB: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("retry: %w", err)
	}

	result := &attachment{
		Reader:     output.JSONReader(att),
		Attachment: att,
	}

	if err := c.fmt.Output(result); err != nil {
		return fmt.Errorf("fmt.Output: %w", err)
	}
	return nil
}

type attachment struct {
	io.Reader
	*kivik.Attachment
}

var _ output.FriendlyOutput = (*attachment)(nil)

func (a *attachment) Execute(w io.Writer) error {
	_, err := io.Copy(w, a.Content)
	_ = a.Content.Close()
	return err
}

func (a *attachment) Read(p []byte) (int, error) {
	n, err := a.Reader.Read(p)
	if err == io.EOF {
		if err := a.Content.Close(); err != nil {
			return n, fmt.Errorf("attachment.Read [EOF]: %w", err)
		}
	}
	if err != nil {
		err = fmt.Errorf("attachment.Read: %w", err)
	}
	return n, err
}
