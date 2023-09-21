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

	"github.com/spf13/cobra"

	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/config"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

type copy struct {
	*root
	target    string
	targetRev string
}

func copyCmd(r *root) *cobra.Command {
	c := &copy{
		root: r,
	}
	cmd := &cobra.Command{
		Use:   "copy [source] [target]",
		Short: "Copy a document",
		Long:  `Copy an existing document.`,
		RunE:  c.RunE,
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&c.target, "target", "t", "", "The target DSN. Useful when reading the source from a config file.")
	pf.StringVarP(&c.targetRev, "target-rev", "R", "", "The current revision of the target document. May also be provided by appending ?rev=<rev> to the target doc ID")

	return cmd
}

func (c *copy) RunE(cmd *cobra.Command, args []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	sourceDB, sourceDoc, err := c.conf.DBDoc()
	if err != nil {
		return err
	}
	if len(args) >= 2 && c.target == "" { // nolint:gomnd
		c.target = args[1]
	}
	if c.target == "" {
		return errors.Code(errors.ErrUsage, "missing target")
	}
	target, targetOpts, err := config.ContextFromDSN(c.target)
	if err != nil {
		return fmt.Errorf("invalid target: %w", err)
	}
	if c.targetRev == "" {
		c.targetRev = targetOpts["rev"]
	}

	c.log.Debugf("[copy] Will copy: %s/%s/%s to %s", client.DSN(), sourceDB, sourceDoc, target.DSN())

	source, _ := c.conf.CurrentCx()
	if !shouldEmulateCopy(source, target) {
		targetDocID := target.DocID
		if c.targetRev != "" {
			targetDocID += "?rev=" + c.targetRev
		}

		return c.retry(func() error {
			rev, err := client.DB(sourceDB).Copy(cmd.Context(), targetDocID, sourceDoc)
			if err != nil {
				return err
			}
			return c.fmt.UpdateResult(target.DocID, rev)
		})
	}

	tClient, err := target.KivikClient(c.parsedConnectTimeout, c.parsedRequestTimeout)
	if err != nil {
		return err
	}

	var doc map[string]interface{}
	var rev kivik.Option
	if c.targetRev != "" {
		rev = kivik.Rev(c.targetRev)
	}

	return c.retry(func() error {
		if doc == nil {
			row := client.DB(sourceDB).Get(cmd.Context(), sourceDoc, c.opts())
			if err := row.Err(); err != nil {
				return err
			}
			if err := row.ScanDoc(&doc); err != nil {
				return err
			}
		}
		rev, err := tClient.DB(target.Database).Put(cmd.Context(), target.DocID, doc, rev)
		if err != nil {
			return err
		}
		return c.fmt.UpdateResult(target.DocID, rev)
	})
}

func shouldEmulateCopy(s, t *config.Context) bool {
	if t.Host != "" && t.Host != s.Host {
		return true
	}
	if t.Database != "" && t.Database != s.Database {
		return true
	}
	return false
}
