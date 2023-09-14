package mockdb

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"gitlab.com/flimzy/testy"

	"github.com/go-kivik/kivik/v4"
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

type multiOptions []kivik.Option

var _ kivik.Option = (multiOptions)(nil)

func (o multiOptions) Apply(t interface{}) {
	for _, opt := range o {
		if opt != nil {
			opt.Apply(t)
		}
	}
}

func Test_convertOptions(t *testing.T) {
	tests := []struct {
		name string
		in   kivik.Option
		want []kivik.Option
	}{
		{
			name: "nil input",
			in:   nil,
			want: nil,
		},
		{
			name: "one items",
			in:   kivik.Rev("x"),
			want: []kivik.Option{kivik.Rev("x")},
		},
		{
			name: "multiOptions",
			in:   multiOptions{kivik.Rev("a"), kivik.Rev("b")},
			want: []kivik.Option{kivik.Rev("a"), kivik.Rev("b")},
		},
		{
			name: "nested multiOptions",
			in: multiOptions{
				multiOptions{kivik.Rev("a"), kivik.Rev("b")},
				multiOptions{kivik.Rev("c"), kivik.Rev("d")},
			},
			want: []kivik.Option{kivik.Rev("a"), kivik.Rev("b"), kivik.Rev("c"), kivik.Rev("d")},
		},
		{
			name: "nil value",
			in:   kivik.Option(nil),
			want: nil,
		},
		{
			name: "filter nil values",
			in: multiOptions{
				multiOptions{kivik.Rev("a"), kivik.Option(nil), kivik.Rev("b"), nil, nil},
				multiOptions{kivik.Option(nil), kivik.Rev("c"), kivik.Rev("d"), kivik.Option(nil)},
			},
			want: []kivik.Option{kivik.Rev("a"), kivik.Rev("b"), kivik.Rev("c"), kivik.Rev("d")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertOptions(tt.in)
			if d := cmp.Diff(tt.want, got); d != "" {
				t.Error(d)
			}
		})
	}
}
