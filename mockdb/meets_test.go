package kivikmock

import (
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestDBMeetsExpectation(t *testing.T) {
	type tst struct {
		exp      *DB
		act      *DB
		expected bool
	}
	tests := testy.NewTable()
	tests.Add("different name", tst{
		exp:      &DB{name: "foo"},
		act:      &DB{name: "bar"},
		expected: false,
	})
	tests.Add("different id", tst{
		exp:      &DB{name: "foo", id: 123},
		act:      &DB{name: "foo", id: 321},
		expected: false,
	})
	tests.Add("no db", tst{
		expected: true,
	})
	tests.Add("match", tst{
		exp:      &DB{name: "foo", id: 123},
		act:      &DB{name: "foo", id: 123},
		expected: true,
	})
	tests.Run(t, func(t *testing.T, test tst) {
		result := dbMeetsExpectation(test.exp, test.act)
		if result != test.expected {
			t.Errorf("Unexpected result: %T", result)
		}
	})
}
