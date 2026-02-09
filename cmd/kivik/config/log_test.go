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

package config

import (
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"

	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4/cmd/kivik/log"
)

// testLogger is a Logger test double. Use the Check() function in tests to
// validate the logs collected.
type testLogger struct {
	mu   sync.Mutex
	logs []string
}

var _ log.Logger = &testLogger{}

// NewTest returns a new test logger.
func newTestLogger() *testLogger {
	return &testLogger{
		logs: []string{},
	}
}

func (l *testLogger) log(level, line string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.logs = append(l.logs, fmt.Sprintf("[%s] %s",
		level,
		strings.TrimSpace(line),
	))
}

func (*testLogger) SetOut(io.Writer) {}
func (*testLogger) SetErr(io.Writer) {}
func (*testLogger) SetDebug(bool)    {}

func (l *testLogger) Debug(args ...any) {
	l.log("DEBUG", fmt.Sprint(args...))
}

func (l *testLogger) Debugf(format string, args ...any) {
	l.log("DEBUG", fmt.Sprintf(format, args...))
}

func (l *testLogger) Info(args ...any) {
	l.log("INFO", fmt.Sprint(args...))
}

func (l *testLogger) Infof(format string, args ...any) {
	l.log("INFO", fmt.Sprintf(format, args...))
}

func (l *testLogger) Error(args ...any) {
	l.log("ERROR", fmt.Sprint(args...))
}

func (l *testLogger) Errorf(format string, args ...any) {
	l.log("ERROR", fmt.Sprintf(format, args...))
}

// Check validates the logs received, against an on-disk golden snapshot.
func (l *testLogger) Check(t *testing.T) {
	t.Helper()
	t.Run("logs", func(t *testing.T) {
		l.mu.Lock()
		t.Cleanup(l.mu.Unlock)
		if d := testy.DiffText(testy.Snapshot(t), strings.Join(l.logs, "\n")); d != nil {
			t.Error(d)
		}
	})
}
