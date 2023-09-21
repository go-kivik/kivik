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

type get struct {
	alldbs, att, doc, db, ver, cf, sec, cluster *cobra.Command
	*root
}

func getCmd(r *root) *cobra.Command {
	g := &get{
		root:    r,
		alldbs:  getAllDBsCmd(r),
		att:     getAttachmentCmd(r),
		doc:     getDocCmd(r),
		db:      getDBCmd(r),
		ver:     getVersionCmd(r),
		cf:      getConfigCmd(r),
		sec:     getSecurityCmd(r),
		cluster: getClusterSetupCmd(r),
	}
	cmd := &cobra.Command{
		Use:   "get [command]",
		Short: "Get a resource",
		Long:  `Fetch a resource described by the URL`,
		RunE:  g.RunE,
	}

	cmd.AddCommand(g.alldbs)
	cmd.AddCommand(g.att)
	cmd.AddCommand(g.doc)
	cmd.AddCommand(g.db)
	cmd.AddCommand(g.ver)
	cmd.AddCommand(g.cf)
	cmd.AddCommand(g.sec)
	cmd.AddCommand(g.cluster)

	return cmd
}

func (g *get) RunE(cmd *cobra.Command, args []string) error {
	dsn, err := g.conf.URL()
	if err != nil {
		return err
	}
	if _, _, ok := configFromDSN(dsn); ok {
		return g.cf.RunE(cmd, args)
	}
	if _, ok := securityFromDSN(dsn); ok {
		return g.sec.RunE(cmd, args)
	}
	if g.conf.HasAttachment() {
		return g.att.RunE(cmd, args)
	}
	if g.conf.HasDoc() {
		return g.doc.RunE(cmd, args)
	}
	if g.conf.HasDB() {
		db, err := g.conf.DB()
		if err != nil {
			return err
		}
		switch db {
		case "_all_dbs":
			return g.alldbs.RunE(cmd, args)
		case "_cluster_setup":
			return g.cluster.RunE(cmd, args)
		}
		return g.db.RunE(cmd, args)
	}
	return g.ver.RunE(cmd, args)
}
