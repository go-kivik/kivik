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

type descrAttachment struct {
	*root
}

func descrAttachmentCmd(r *root) *cobra.Command {
	c := &descrAttachment{
		root: r,
	}
	return &cobra.Command{
		Use:     "attachment [dsn]/[database]/[document]/[filename]",
		Aliases: []string{"att", "attach"},
		Short:   "Describe an attachment",
		Long:    `Fetch attachment with the HTTP HEAD verb`,
		RunE:    c.RunE,
	}
}

func (c *descrAttachment) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	db, docID, filename, err := c.conf.DBDocFilename()
	if err != nil {
		return err
	}
	c.log.Debugf("[get] Will fetch document: %s/%s/%s", client.DSN(), db, docID)
	return c.retry(func() error {
		att, err := client.DB(db).GetAttachmentMeta(cmd.Context(), docID, filename, c.opts())
		if err != nil {
			return err
		}
		att.Filename = filename

		format := `Filename: {{ .Filename }}
Content-Type: {{ .ContentType }}
Content-Length: {{ .Size }}
Digest: {{ .Digest }}
{{- with .RevPos }}
RevPos: {{ . }}
{{ end -}}
`

		result := output.TemplateReader(format, att, output.JSONReader(att))
		return c.fmt.Output(result)
	})
}
