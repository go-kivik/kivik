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

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"gitlab.com/flimzy/testy"
)

// TestLogger is a Logger test double. Use the Check() function in tests to
// validate the logs collected.
type TestLogger struct {
	mu   sync.Mutex
	logs []string
}

var _ Logger = &TestLogger{}

// NewTest returns a new test logger.
func NewTest() *TestLogger {
	return &TestLogger{
		logs: []string{},
	}
}

func (l *TestLogger) log(level, line string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, fmt.Sprintf("[%s] %s",
		level,
		strings.TrimSpace(line),
	))
}

func (*TestLogger) SetOut(io.Writer) {}
func (*TestLogger) SetErr(io.Writer) {}
func (*TestLogger) SetDebug(bool)    {}

func (l *TestLogger) Debug(args ...interface{}) {
	l.log("DEBUG", fmt.Sprint(args...))
}

func (l *TestLogger) Debugf(format string, args ...interface{}) {
	l.log("DEBUG", fmt.Sprintf(format, args...))
}

func (l *TestLogger) Info(args ...interface{}) {
	l.log("INFO", fmt.Sprint(args...))
}

func (l *TestLogger) Infof(format string, args ...interface{}) {
	l.log("INFO", fmt.Sprintf(format, args...))
}

func (l *TestLogger) Error(args ...interface{}) {
	l.log("ERROR", fmt.Sprint(args...))
}

func (l *TestLogger) Errorf(format string, args ...interface{}) {
	l.log("ERROR", fmt.Sprintf(format, args...))
}

// Check validates the logs received, against an on-disk golden snapshot.
func (l *TestLogger) Check(t *testing.T) {
	t.Helper()
	t.Run("logs", func(t *testing.T) {
		l.mu.Lock()
		defer l.mu.Unlock()
		if d := testy.DiffText(testy.Snapshot(t), strings.Join(l.logs, "\n")); d != nil {
			t.Error(d)
		}
	})
}
