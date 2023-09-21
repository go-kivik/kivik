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
	"net/http"

	"github.com/spf13/cobra"

	"github.com/go-kivik/kivik/v4/couchdb/chttp"

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
)

type ping struct {
	*root
}

func pingCmd(r *root) *cobra.Command {
	c := &ping{
		root: r,
	}

	return &cobra.Command{
		Use:   "ping [dsn]",
		Short: "Ping a server",
		Long:  "Ping a server's /_up endpoint to determine availability to serve requests",
		RunE:  c.RunE,
	}
}

func (c *ping) RunE(cmd *cobra.Command, args []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	c.conf.Finalize()
	c.log.Debugf("[ping] Will ping server: %q", client.DSN())
	return c.retry(func() error {
		var status int
		ctx := chttp.WithClientTrace(cmd.Context(), &chttp.ClientTrace{
			HTTPResponse: func(res *http.Response) {
				status = res.StatusCode
			},
		})
		success, err := client.Ping(ctx)
		if err != nil {
			return err
		}
		if success {
			c.log.Info("[ping] Server is up")
			return nil
		}
		c.log.Info("[ping] Server down")
		return errors.HTTPStatus(status, "Server down")
	})
}
