package main

import (
	"reflect"
	"testing"

	"gitlab.com/flimzy/testy"
)

func init() {
	initTemplates("templates")
}

func TestRenderExpectedType(t *testing.T) {
	tests := testy.NewTable()
	tests.Add("CreateDoc", &method{
		Name:                 "CreateDoc",
		DBMethod:             true,
		AcceptsContext:       true,
		AcceptsLegacyOptions: true,
		ReturnsError:         true,
		Accepts:              []reflect.Type{reflect.TypeOf((*interface{})(nil)).Elem()},
		Returns:              []reflect.Type{typeString, typeString},
	})

	tests.Run(t, func(t *testing.T, m *method) {
		result, err := renderExpectedType(m)
		if err != nil {
			t.Fatal(err)
		}
		if d := testy.DiffText(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

func TestRenderDriverMethod(t *testing.T) {
	type tst struct {
		method *method
		err    string
	}
	tests := testy.NewTable()
	tests.Add("CreateDB", tst{
		method: &method{
			Name:                 "CreateDB",
			Accepts:              []reflect.Type{typeString},
			AcceptsContext:       true,
			AcceptsLegacyOptions: true,
			ReturnsError:         true,
		},
	})
	tests.Add("No context", tst{
		method: &method{
			Name:                 "NoCtx",
			AcceptsLegacyOptions: true,
			ReturnsError:         true,
		},
	})
	tests.Run(t, func(t *testing.T, test tst) {
		result, err := renderDriverMethod(test.method)
		testy.Error(t, test.err, err)
		if d := testy.DiffText(testy.Snapshot(t), result); d != nil {
			t.Error(d)
		}
	})
}

func TestVariables(t *testing.T) {
	type tst struct {
		method   *method
		indent   int
		expected string
	}
	tests := testy.NewTable()
	tests.Add("no args", tst{
		method:   &method{},
		expected: "",
	})
	tests.Add("one arg", tst{
		method:   &method{Accepts: []reflect.Type{typeString}},
		expected: "arg0: arg0,",
	})
	tests.Add("one arg + options", tst{
		method: &method{Accepts: []reflect.Type{typeString}, AcceptsLegacyOptions: true},
		expected: `arg0:    arg0,
options: options,`,
	})
	tests.Add("indent", tst{
		method: &method{Accepts: []reflect.Type{typeString, typeString}, AcceptsLegacyOptions: true},
		indent: 2,
		expected: `		arg0:    arg0,
		arg1:    arg1,
		options: options,`,
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result := test.method.Variables(test.indent)
		if d := testy.DiffText(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}

// func TestQuotedZero(t *testing.T) {
// 	type tst struct {
// 		input    reflect.Type
// 		expected string
// 	}
// 	tests := testy.NewTable()
// 	tests.Add("string", tst{
// 		input:    reflect.TypeOf(""),
// 		expected: `""`,
// 	})
// 	tests.Add("[]string", tst{
// 		input:    reflect.TypeOf([]string{}),
// 		expected: `[]string(nil)`,
// 	})
//
// 	tests.Run(t, func(t *testing.T, test tst) {
// 		result := quotedZero(test.input)
// 		if result != test.expected {
// 			t.Errorf("Unexpected return: %s", result)
// 		}
// 	})
// }
