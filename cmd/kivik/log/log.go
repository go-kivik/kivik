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

// Package log handles logging.
package log

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// Logger is the standard logger interface.
type Logger interface {
	// SetOut sets the destination for normal output.
	SetOut(io.Writer)
	// SetErr sets the destination for error output.
	SetErr(io.Writer)
	// SetDebug turns debug mode on or off.
	SetDebug(bool)
	// Debug logs debug output.
	Debug(...any)
	// Debug logs formatted debug output.
	Debugf(string, ...any)
	// Info logs normal priority messages.
	Info(...any)
	// Infof logs formatted normal priority messages.
	Infof(string, ...any)
	// Error logs error messages.
	Error(...any)
	// Errorf logs formatted error messages.
	Errorf(string, ...any)
}

type logger struct {
	stdout io.Writer
	stderr io.Writer
	debug  bool
}

var _ Logger = &logger{}

// New returns a new logger instance.
func New() Logger {
	return &logger{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (l *logger) SetOut(out io.Writer) { l.stdout = out }
func (l *logger) SetErr(err io.Writer) { l.stderr = err }
func (l *logger) SetDebug(debug bool)  { l.debug = debug }

func (l *logger) err(line string) {
	_, _ = fmt.Fprintln(l.stderr, strings.TrimSpace(line))
}

func (l *logger) out(line string) {
	_, _ = fmt.Fprintln(l.stdout, strings.TrimSpace(line))
}

func (l *logger) Debug(args ...any) {
	if l.debug {
		l.err(fmt.Sprint(args...))
	}
}

func (l *logger) Debugf(format string, args ...any) {
	if l.debug {
		l.err(fmt.Sprintf(format, args...))
	}
}

func (l *logger) Info(args ...any) {
	l.out(fmt.Sprint(args...))
}

func (l *logger) Infof(format string, args ...any) {
	l.out(fmt.Sprintf(format, args...))
}

func (l *logger) Error(args ...any) {
	l.err(fmt.Sprint(args...))
}

func (l *logger) Errorf(format string, args ...any) {
	l.err(fmt.Sprintf(format, args...))
}
