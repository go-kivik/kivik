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

//go:build go1.18
// +build go1.18

package kiviktest

import (
	"io"
	"os"
	"reflect"
	"regexp"
	"testing"
	"time"
)

type corpusEntry = struct {
	Parent     string
	Path       string
	Data       []byte
	Values     []interface{}
	Generation int
	IsSeed     bool
}

// testDeps is a copy of testing.testDeps
type testDeps interface {
	ImportPath() string
	MatchString(pat, str string) (bool, error)
	SetPanicOnExit0(bool)
	StartCPUProfile(io.Writer) error
	StopCPUProfile()
	StartTestLog(io.Writer)
	StopTestLog() error
	WriteProfileTo(string, io.Writer, int) error
	CoordinateFuzzing(time.Duration, int64, time.Duration, int64, int, []corpusEntry, []reflect.Type, string, string) error
	RunFuzzWorker(func(corpusEntry) error) error
	ReadCorpus(string, []reflect.Type) ([]corpusEntry, error)
	CheckCorpus([]interface{}, []reflect.Type) error
	ResetCoverage()
	SnapshotCoverage()
}

type deps struct{}

var _ testDeps = &deps{}

func (*deps) MatchString(pat, str string) (bool, error)         { return regexp.MatchString(pat, str) }
func (*deps) StartCPUProfile(_ io.Writer) error                 { return nil }
func (*deps) StopCPUProfile()                                   {}
func (*deps) WriteHeapProfile(_ io.Writer) error                { return nil }
func (*deps) WriteProfileTo(_ string, _ io.Writer, _ int) error { return nil }
func (*deps) ImportPath() string                                { return "" }
func (*deps) StartTestLog(io.Writer)                            {}
func (*deps) StopTestLog() error                                { return nil }
func (*deps) SetPanicOnExit0(bool)                              {}
func (*deps) CheckCorpus([]interface{}, []reflect.Type) error   { return nil }
func (*deps) CoordinateFuzzing(time.Duration, int64, time.Duration, int64, int, []corpusEntry, []reflect.Type, string, string) error {
	return nil
}
func (*deps) RunFuzzWorker(func(corpusEntry) error) error              { return nil }
func (*deps) ReadCorpus(string, []reflect.Type) ([]corpusEntry, error) { return nil, nil }
func (*deps) ResetCoverage()                                           {}
func (*deps) SnapshotCoverage()                                        {}

func mainStart(tests []testing.InternalTest) {
	m := testing.MainStart(&deps{}, tests, nil, nil, nil)
	os.Exit(m.Run())
}
