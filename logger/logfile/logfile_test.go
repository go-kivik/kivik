package logfile

import (
	"bytes"
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/flimzy/kivik/logger"
)

type logTest struct {
	Name      string
	Messages  []string
	BufSize   int64
	ExpectedN int64
	Expected  string
}

var CTX = context.Background()

func TestLogErrors(t *testing.T) {
	f, err := ioutil.TempFile("", "kivik-log-")
	if err != nil {
		t.Errorf("Failed to create temp file: %s", err)
		return
	}
	defer os.Remove(f.Name())
	client := &Logger{}
	if err = client.Init(map[string]string{"file": f.Name()}); err != nil {
		t.Errorf("Failed to open logger: %s", err)
	}

	if _, err = client.LogContext(CTX, -100, 0); err == nil {
		t.Errorf("No error for invalid length argument")
	}
	if _, err = client.LogContext(CTX, 0, -100); err == nil {
		t.Errorf("No error for invalid offset argument")
	}

	log, err := client.LogContext(CTX, 0, 0)
	if err != nil {
		t.Errorf("Error reading 0 bytes: %s", err)
	}
	defer log.Close()
	buf := &bytes.Buffer{}
	buf.ReadFrom(log)
	if buf.Len() > 0 {
		t.Errorf("Read more than 0 bytes")
	}
}

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
			BufSize:  100,
			Expected: "[Mon, 02 Jan 2006 15:04:05 MST] [info] [--] test 1\n",
		},
		logTest{
			Name:     "Two logs",
			Messages: []string{"test 1", "test 2"},
			BufSize:  100,
			Expected: "on, 02 Jan 2006 15:04:05 MST] [info] [--] test 1\n[Mon, 02 Jan 2006 15:04:05 MST] [info] [--] test 2\n",
		},
		logTest{
			Name:     "Overflow ring",
			Messages: []string{"test 1", "test 2", "test 3", "Abracadabra"},
			BufSize:  100,
			Expected: "2 Jan 2006 15:04:05 MST] [info] [--] test 3\n[Mon, 02 Jan 2006 15:04:05 MST] [info] [--] Abracadabra\n",
		},
	}
	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			f, err := ioutil.TempFile("", "kivik-log-")
			if err != nil {
				t.Errorf("Failed to create temp file: %s", err)
				return
			}
			defer os.Remove(f.Name())
			log := &Logger{}
			if err = log.Init(map[string]string{"file": f.Name()}); err != nil {
				t.Errorf("Failed to open logger: %s", err)
			}
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
