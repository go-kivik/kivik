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
)

type describe struct {
	att, doc, db, ver *cobra.Command
	*root
}

func descrCmd(r *root) *cobra.Command {
	g := &describe{
		root: r,
		att:  descrAttachmentCmd(r),
		doc:  descrDocCmd(r),
		db:   descrDBCmd(r),
		ver:  descrVerCmd(r),
	}
	cmd := &cobra.Command{
		Use:     "describe [command]",
		Aliases: []string{"desc", "descr"},
		Short:   "Describe a resource",
		Long:    `Describe a resource described by the URL`,
		RunE:    g.RunE,
	}

	cmd.AddCommand(g.att)
	cmd.AddCommand(g.doc)
	cmd.AddCommand(g.db)
	cmd.AddCommand(g.ver)

	return cmd
}

func (g *describe) RunE(cmd *cobra.Command, args []string) error {
	if g.conf.HasAttachment() {
		return g.att.RunE(cmd, args)
	}
	if g.conf.HasDoc() {
		return g.doc.RunE(cmd, args)
	}
	if g.conf.HasDB() {
		return g.db.RunE(cmd, args)
	}
	return g.ver.RunE(cmd, args)
}
