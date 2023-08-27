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

// +build go1.8,!go1.9

package kiviktest

import (
	"io"
	"os"
	"regexp"
	"testing"
)

/* This file contains copies of basic functionality from the testing package */

// testDeps is a copy of testing.testDeps
type testDeps interface {
	MatchString(pat, str string) (bool, error)
	StartCPUProfile(io.Writer) error
	StopCPUProfile()
	WriteHeapProfile(io.Writer) error
	WriteProfileTo(string, io.Writer, int) error
}

type deps struct{}

var _ testDeps = &deps{}

func (d *deps) MatchString(pat, str string) (bool, error)         { return regexp.MatchString(pat, str) }
func (d *deps) StartCPUProfile(_ io.Writer) error                 { return nil }
func (d *deps) StopCPUProfile()                                   {}
func (d *deps) WriteHeapProfile(_ io.Writer) error                { return nil }
func (d *deps) WriteProfileTo(_ string, _ io.Writer, _ int) error { return nil }

func mainStart(tests []testing.InternalTest) {
	m := testing.MainStart(&deps{}, tests, nil, nil)
	os.Exit(m.Run())
}
