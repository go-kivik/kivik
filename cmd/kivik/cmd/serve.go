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

// import (
// 	"fmt"

// 	"github.com/spf13/cobra"
// )

// const defaultCouchDBPort = 5984

// type serve struct {
// 	*root
// 	port int
// }

// func serveCmd(r *root) *cobra.Command {
// 	s := &serve{
// 		root: r,
// 	}

// 	cmd := &cobra.Command{
// 		Use:   "serve [dsn]",
// 		Short: "[EXPERIMENTAL] Start HTTP server",
// 		Long:  "[EXPERIMENTAL] Serves the resources located at the specified DSN via HTTP.",
// 		RunE:  s.RunE,
// 	}

// 	f := cmd.Flags()
// 	f.IntVarP(&s.port, "port", "p", defaultCouchDBPort, "HTTP port to listen on")

// 	return cmd
// }

// func (s *serve) RunE(*cobra.Command, []string) error {
// 	fmt.Printf("port: %d\n", s.port)
// 	return nil
// }
