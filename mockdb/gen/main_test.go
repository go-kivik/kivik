package main

import (
	"testing"

	"gitlab.com/flimzy/testy"
)

func TestCompareMethods(t *testing.T) {
	type tst struct {
		client    []*method
		driver    []*method
		expSame   []*method
		expClient []*method
		expDriver []*method
	}
	tests := testy.NewTable()
	tests.Add("one identical", tst{
		client: []*method{
			{Name: "Foo"},
		},
		driver: []*method{
			{Name: "Foo"},
		},
		expSame: []*method{
			{Name: "Foo"},
		},
		expClient: []*method{},
		expDriver: []*method{},
	})
	tests.Add("same name", tst{
		client: []*method{
			{Name: "Foo", ReturnsError: true},
		},
		driver: []*method{
			{Name: "Foo"},
		},
		expSame: []*method{},
		expClient: []*method{
			{Name: "Foo", ReturnsError: true},
		},
		expDriver: []*method{
			{Name: "Foo"},
		},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		same, client, driver := compareMethods(test.client, test.driver)
		if d := testy.DiffInterface(test.expSame, same); d != nil {
			t.Errorf("Same:\n%s\n", d)
		}
		if d := testy.DiffInterface(test.expClient, client); d != nil {
			t.Errorf("Same:\n%s\n", d)
		}
		if d := testy.DiffInterface(test.expDriver, driver); d != nil {
			t.Errorf("Same:\n%s\n", d)
		}
	})
}
