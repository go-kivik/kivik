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

package log

import "io"

type nilLogger struct{}

var _ Logger = nilLogger{}

// NewNil returns a new nil logger. All logs are silently discarded.
func NewNil() Logger { return nilLogger{} }

func (nilLogger) SetOut(io.Writer)              {}
func (nilLogger) SetErr(io.Writer)              {}
func (nilLogger) SetDebug(bool)                 {}
func (nilLogger) Debug(...interface{})          {}
func (nilLogger) Debugf(string, ...interface{}) {}
func (nilLogger) Info(...interface{})           {}
func (nilLogger) Infof(string, ...interface{})  {}
func (nilLogger) Error(...interface{})          {}
func (nilLogger) Errorf(string, ...interface{}) {}
