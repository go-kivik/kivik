package kivik

import (
	"testing"
	"time"
)

var testOptions = map[string]interface{}{"foo": 123}

func parseTime(t *testing.T, str string) time.Time {
	ts, err := time.Parse(time.RFC3339, str)
	if err != nil {
		t.Fatal(err)
	}
	return ts
}
