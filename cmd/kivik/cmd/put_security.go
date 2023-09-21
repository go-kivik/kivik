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

type putSec struct {
	*root
	*input.Input
}

func putSecurityCmd(p *put) *cobra.Command {
	c := &putSec{
		root:  p.root,
		Input: p.Input,
	}
	cmd := &cobra.Command{
		Use:   "security [dsn]/[database]",
		Short: "Set database security object",
		RunE:  c.RunE,
	}

	return cmd
}

func (c *putSec) RunE(cmd *cobra.Command, _ []string) error {
	client, err := c.client()
	if err != nil {
		return err
	}
	dsn, err := c.conf.URL()
	if err != nil {
		return err
	}
	db, ok := securityFromDSN(dsn)
	if !ok {
		db, err = c.conf.DB()
		if err != nil {
			return err
		}
	}

	secObj := new(kivik.Security)
	if err := c.As(&secObj); err != nil {
		return err
	}
	c.log.Debugf("[get] Will put security object: %s/%s", client.DSN(), db)

	return c.retry(func() error {
		err = client.DB(db).SetSecurity(cmd.Context(), secObj)
		if err != nil {
			return err
		}

		return c.fmt.OK()
	})
}
