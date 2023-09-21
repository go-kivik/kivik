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

//go:build !js
// +build !js

package test

import (
	"os"
	"testing"

	_ "github.com/go-kivik/fsdb/v4"
	"github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/kiviktest"
	"github.com/go-kivik/kivik/v4/kiviktest/kt"
)

func init() {
	RegisterFSSuite()
}

func TestFS(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "kivik.test.")
	if err != nil {
		t.Errorf("Failed to create temp dir to test FS driver: %s\n", err)
		return
	}
	defer os.RemoveAll(tempDir) //nolint:errcheck // safe to ignore
	client, err := kivik.New("fs", tempDir)
	if err != nil {
		t.Errorf("Failed to connect to FS driver: %s\n", err)
		return
	}
	clients := &kt.Context{
		RW:    true,
		Admin: client,
		T:     t,
	}
	kiviktest.RunTestsInternal(clients, kiviktest.SuiteKivikFS)
}
