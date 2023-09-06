package main

import (
	"context"
	"reflect"
	"testing"

	"gitlab.com/flimzy/testy"

	kivik "github.com/go-kivik/kivik/v4"
)

type testDriver interface {
	WithCtx(context.Context) error
	NoCtx(string) error
	WithOptions(string, map[string]interface{})
}

type testClient struct{}

func (c *testClient) WithCtx(_ context.Context) error          { return nil }
func (c *testClient) NoCtx(_ string) error                     { return nil }
func (c *testClient) WithOptions(_ string, _ ...kivik.Options) {}

func TestMethods(t *testing.T) {
	type tst struct {
		input    interface{}
		isClient bool
		expected []*Method
		err      string
	}
	tests := testy.NewTable()
	tests.Add("non-struct", tst{
		input: 123,
		err:   "input must be struct",
	})
	tests.Add("wrong field name", tst{
		input: struct{ Y int }{}, // nolint: unused
		err:   "wrapper struct must have a single field: X",
	})
	tests.Add("non-interface", tst{
		input: struct{ X int }{}, // nolint: unused
		err:   "field X must be of type interface",
	})
	tests.Add("testDriver", tst{
		input: struct{ X testDriver }{}, // nolint: unused
		expected: []*Method{
			{
				Name:         "NoCtx",
				ReturnsError: true,
				Accepts:      []reflect.Type{typeString},
			},
			{
				Name:           "WithCtx",
				AcceptsContext: true,
				ReturnsError:   true,
			},
			{
				Name:           "WithOptions",
				AcceptsOptions: true,
				Accepts:        []reflect.Type{typeString},
			},
		},
	})
	tests.Add("invalid client", tst{
		input:    struct{ X int }{}, // nolint: unused
		isClient: true,
		err:      "field X must be of type pointer to struct",
	})
	tests.Add("testClient", tst{
		input:    struct{ X testClient }{}, // nolint: unused
		isClient: true,
		err:      "field X must be of type pointer to struct",
	})
	tests.Add("*testClient", tst{
		input:    struct{ X *testClient }{}, // nolint: unused
		isClient: true,
		expected: []*Method{
			{
				Name:         "NoCtx",
				ReturnsError: true,
				Accepts:      []reflect.Type{typeString},
			},
			{
				Name:           "WithCtx",
				AcceptsContext: true,
				ReturnsError:   true,
			},
			{
				Name:           "WithOptions",
				AcceptsOptions: true,
				Accepts:        []reflect.Type{typeString},
			},
		},
	})

	tests.Run(t, func(t *testing.T, test tst) {
		result, err := parseMethods(test.input, test.isClient, nil)
		testy.Error(t, test.err, err)
		if d := testy.DiffInterface(test.expected, result); d != nil {
			t.Error(d)
		}
	})
}
