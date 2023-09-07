// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

// Package main launches test containers and runs tests.
package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
)

func main() {
	const commandIndex = 2
	if len(os.Args) < commandIndex+1 {
		usage()
	}
	command := os.Args[commandIndex:]
	args := make([]string, len(command)-1)
	if len(command) > 1 {
		copy(args, command[1:])
	}
	cmd := exec.CommandContext(context.TODO(), command[0], args...)
	cmd.Stderr = os.Stderr
	cmd.Stdout = os.Stdout
	if err := cmd.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `USAGE:

go run github.com/go-kivik/kivik/v4/internal/testcontainers <suite[,suite...]> <command>
`)
	os.Exit(1)
}
