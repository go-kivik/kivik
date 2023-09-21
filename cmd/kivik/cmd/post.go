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

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/input"
)

type post struct {
	*root
	*input.Input
	doc, vc, flush, compact, cv, purge, repl, cluster *cobra.Command
}

func postCmd(r *root) *cobra.Command {
	c := &post{
		root:    r,
		Input:   input.New(),
		vc:      postViewCleanupCmd(r),
		flush:   postFlushCmd(r),
		compact: postCompactCmd(r),
		cv:      postCompactViewsCmd(r),
		repl:    postReplicateCmd(r),
	}
	c.doc = postDocCmd(c)
	c.purge = postPurgeCmd(c)
	c.cluster = postClusterSetupCmd(c)

	cmd := &cobra.Command{
		Use:   "post",
		Short: "Post a resource",
		Long:  `Post to the named resource`,
		RunE:  c.RunE,
	}

	c.Input.ConfigFlags(cmd.PersistentFlags())

	cmd.AddCommand(c.doc)
	cmd.AddCommand(c.vc)
	cmd.AddCommand(c.flush)
	cmd.AddCommand(c.compact)
	cmd.AddCommand(c.cv)
	cmd.AddCommand(c.purge)
	cmd.AddCommand(c.repl)
	cmd.AddCommand(c.cluster)

	return cmd
}

func dbCommandFromDSN(dsn *url.URL) (command, db string) {
	parts := strings.Split(dsn.Path, "/")
	if len(parts) != 3 { // nolint:gomnd
		return "", ""
	}
	return parts[2], parts[1]
}

func (c *post) RunE(cmd *cobra.Command, args []string) error {
	dsn, err := c.conf.URL()
	if err != nil {
		return err
	}
	if db, _ := compactViewFromDSN(dsn); db != "" {
		return c.cv.RunE(cmd, args)
	}
	switch command, _ := dbCommandFromDSN(dsn); command {
	case "_view_cleanup":
		return c.vc.RunE(cmd, args)
	case "_ensure_full_commit":
		return c.flush.RunE(cmd, args)
	case "_compact":
		return c.compact.RunE(cmd, args)
	case "_purge":
		return c.purge.RunE(cmd, args)
	}
	switch dsn.Path {
	case "/_replicate":
		return c.repl.RunE(cmd, args)
	case "/_cluster_setup":
		return c.cluster.RunE(cmd, args)
	}
	if c.conf.HasDB() {
		return c.doc.RunE(cmd, args)
	}
	_, err = c.client()
	if err != nil {
		return err
	}

	return errors.New("xxx")
}
