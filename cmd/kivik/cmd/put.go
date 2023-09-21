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

	"github.com/go-kivik/xkivik/v4/cmd/kivik/errors"
	"github.com/go-kivik/xkivik/v4/cmd/kivik/input"
)

type put struct {
	*input.Input
	*root

	db, doc, att, cf, sec *cobra.Command
}

func putCmd(r *root) *cobra.Command {
	c := &put{
		root:  r,
		Input: input.New(),
		db:    putDBCmd(r),
	}
	c.doc = putDocCmd(c)
	c.att = putAttCmd(c)
	c.cf = putConfigCmd(c)
	c.sec = putSecurityCmd(c)

	cmd := &cobra.Command{
		Use:   "put",
		Short: "Put a resource",
		Long:  `Create or update the named resource`,
		RunE:  c.RunE,
	}

	c.Input.ConfigFlags(cmd.PersistentFlags())

	cmd.AddCommand(c.db)
	cmd.AddCommand(c.doc)
	cmd.AddCommand(c.att)
	cmd.AddCommand(c.cf)
	cmd.AddCommand(c.sec)

	return cmd
}

func (c *put) RunE(cmd *cobra.Command, args []string) error {
	dsn, err := c.conf.URL()
	if err != nil {
		return err
	}
	if _, _, ok := configFromDSN(dsn); ok {
		return c.cf.RunE(cmd, args)
	}
	if _, ok := securityFromDSN(dsn); ok {
		return c.sec.RunE(cmd, args)
	}
	if c.conf.HasAttachment() {
		return c.att.RunE(cmd, args)
	}
	if c.conf.HasDoc() {
		return c.doc.RunE(cmd, args)
	}
	if c.conf.HasDB() {
		return c.db.RunE(cmd, args)
	}
	_, err = c.client()
	if err != nil {
		return err
	}

	return errors.New("xxx")
}
