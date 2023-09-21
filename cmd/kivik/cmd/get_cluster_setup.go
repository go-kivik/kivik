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

type getClusterSetup struct {
	*root
}

func getClusterSetupCmd(r *root) *cobra.Command {
	g := &getClusterSetup{
		root: r,
	}
	return &cobra.Command{
		Use:     "cluster-setup [dsn]",
		Aliases: []string{"cluster"},
		Short:   "Get the status of the node or cluster",
		RunE:    g.RunE,
	}
}

func (c *getClusterSetup) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}

	c.log.Debugf("[get] Will fetch cluster setup: %s", client.DSN())
	return c.retry(func() error {
		state, err := client.ClusterStatus(cmd.Context())
		if err != nil {
			return err
		}
		data := struct {
			State         string `json:"state"`
			FriendlyState string
		}{
			State:         state,
			FriendlyState: friendlyState(state),
		}
		format := `Cluster status: {{.FriendlyState}}`
		result := output.TemplateReader(format, data, output.JSONReader(map[string]string{"state": state}))
		return c.fmt.Output(result)
	})
}

func friendlyState(state string) string {
	switch state {
	case "cluster_disabled":
		return "Cluster disabled"
	case "single_node_disabled":
		return "Single node disabled"
	case "single_node_enabled":
		return "Single node enabled"
	case "cluster_enabled":
		return "Cluster enabled"
	case "cluster_finished":
		return "Cluster finished"
	}
	return state
}
