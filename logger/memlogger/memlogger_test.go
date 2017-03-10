package memlogger

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/flimzy/kivik/logger"
)

type logTest struct {
	Name      string
	Messages  []string
	RingSize  int
	BufSize   int64
	ExpectedN int
	Expected  string
}

var CTX = context.Background()

func TestLog(t *testing.T) {
	now = func() time.Time {
		n, err := time.Parse("Mon Jan 2 15:04:05 -0700 MST 2006", "Mon Jan 2 15:04:05 -0700 MST 2006")
		if err != nil {
			panic(err)
		}
		return n
	}
	tests := []logTest{
		logTest{
			Name:     "Single long",
			Messages: []string{"test 1"},
			RingSize: 2,
			BufSize:  100,
			Expected: "[Mon, 02 Jan 2006 15:04:05 MST] [info] [--] test 1\n",
		},
		logTest{
			Name:     "Two logs",
			Messages: []string{"test 1", "test 2"},
			RingSize: 2,
			BufSize:  100,
			Expected: "on, 02 Jan 2006 15:04:05 MST] [info] [--] test 1\n[Mon, 02 Jan 2006 15:04:05 MST] [info] [--] test 2\n",
		},
		logTest{
			Name:     "Overflow ring",
			Messages: []string{"test 1", "test 2", "test 3", "Abracadabra"},
			RingSize: 2,
			BufSize:  100,
			Expected: "2 Jan 2006 15:04:05 MST] [info] [--] test 3\n[Mon, 02 Jan 2006 15:04:05 MST] [info] [--] Abracadabra\n",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			log := &Logger{}
			log.Init(map[string]string{
				"capacity": fmt.Sprintf("%d", test.RingSize),
			})
			for _, msg := range test.Messages {
				log.WriteLog(logger.LogLevelInfo, msg)
			}
			logR, err := log.LogContext(CTX, test.BufSize, 0)
			if err != nil {
				t.Fatalf("Unexpected error reading log: %s", err)
			}
			defer logR.Close()
			buf := &bytes.Buffer{}
			buf.ReadFrom(logR)
			if buf.Len() != len(test.Expected) {
				t.Errorf("Expected to read %d bytes, but read %d\n", len(test.Expected), buf.Len())
			}
			if string(buf.String()) != test.Expected {
				t.Errorf("Logs don't match\nExpected: %s\n  Acutal: %s\n", test.Expected, buf.String())
			}
		})
	}
}
