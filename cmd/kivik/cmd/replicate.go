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
	"strings"

	"github.com/spf13/cobra"

	"github.com/go-kivik/kivik/v4"
	_ "github.com/go-kivik/kivik/v4/x/fsdb" // Filesystem driver

	"github.com/go-kivik/xkivik/v4"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/config"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/output"
)

type replicate struct {
	*root
}

func replicateCmd(r *root) *cobra.Command {
	c := &replicate{
		root: r,
	}

	cmd := &cobra.Command{
		Use:   "replicate",
		Short: "Replicate a database",
		Long: `Replicate source to target, managed by couchctl. For a server-controlled replication, use the 'post replicate' command instead.

This command supports a limited version of the CouchDB replication protocol. The following options are supported:

filter (string) - The name of a filter function.
doc_ids (array of string) - Array of document IDs to be synchronized.
copy_security (bool) - When true, the security object is read from the source, and copied to the target, before the replication. Use with caution! The security object is not versioned, and will be unconditionally overwritten!`,
		RunE: c.RunE,
	}

	return cmd
}

func (c *replicate) connect(key string) (*kivik.DB, error) {
	opts := c.options
	dsn, _ := opts[key].(string)
	if dsn == "" {
		return nil, errors.Codef(errors.ErrUsage, "missing %s", key)
	}

	// Special case for relative files, since the DSN doesn't need to represent
	// a "server"
	if dsn[0] == '.' || dsn[0] == '/' || strings.HasPrefix(dsn, "file://") {
		client, err := kivik.New("fs", "")
		if err != nil {
			return nil, fmt.Errorf("%s: %w", key, err)
		}
		db := client.DB(dsn)
		return db, db.Err()
	}
	cx, _, err := config.ContextFromDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	db, err := cx.DB()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	client, err := cx.KivikClient(c.parsedConnectTimeout, c.parsedRequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", key, err)
	}
	return client.DB(db), nil
}

func (c *replicate) RunE(cmd *cobra.Command, args []string) error {
	c.conf.Finalize()
	source, err := c.connect("source")
	if err != nil {
		return err
	}
	target, err := c.connect("target")
	if err != nil {
		return err
	}

	opts := c.options
	c.log.Debugf("[replicate] Will replicate %s to %s", opts["source"], opts["target"])
	result, err := xkivik.Replicate(cmd.Context(), target, source, kivik.Params(opts))
	if err != nil {
		return err
	}
	return c.fmt.Output(output.JSONReader(result))
}
