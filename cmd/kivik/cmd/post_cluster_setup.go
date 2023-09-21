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

type postClusterSetup struct {
	*root
	*input.Input
}

func postClusterSetupCmd(p *post) *cobra.Command {
	c := &postClusterSetup{
		root:  p.root,
		Input: p.Input,
	}
	cmd := &cobra.Command{
		Use:     "cluster-setup [dsn]",
		Aliases: []string{"cluster"},
		Short:   "Configure node as standalone node or finalize a cluster",
		RunE:    c.RunE,
	}

	return cmd
}

func (c *postClusterSetup) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}

	data, err := c.JSONData()
	if err != nil {
		return err
	}
	c.log.Debugf("[get] Will put cluster setup object: %s", client.DSN())

	return c.retry(func() error {
		err = client.ClusterSetup(cmd.Context(), data)
		if err != nil {
			return err
		}

		return c.fmt.OK()
	})
}
