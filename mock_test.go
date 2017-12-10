package kivik

import (
	"context"
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/flimzy/kivik/driver"
)

type mockDB struct {
	driver.DB
	ChangesFunc func(context.Context, map[string]interface{}) (driver.Changes, error)
}

var _ driver.DB = &mockDB{}

func (db *mockDB) Changes(ctx context.Context, opts map[string]interface{}) (driver.Changes, error) {
	return db.ChangesFunc(ctx, opts)
}

type mockExplainer struct {
	driver.DB
	plan *driver.QueryPlan
	err  error
}

var _ driver.Explainer = &mockExplainer{}

func (db *mockExplainer) Explain(_ context.Context, query interface{}) (*driver.QueryPlan, error) {
	return db.plan, db.err
}

type errReader string

var _ io.ReadCloser = errReader("")

func (r errReader) Close() error {
	return nil
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, errors.New(string(r))
}

type mockBulkResults struct {
	result *driver.BulkResult
	err    error
}

var _ driver.BulkResults = &mockBulkResults{}

func (r *mockBulkResults) Next(i *driver.BulkResult) error {
	if r.result != nil {
		*i = *r.result
	}
	return r.err
}

func (r *mockBulkResults) Close() error { return nil }

func body(s string) io.ReadCloser {
	return ioutil.NopCloser(strings.NewReader(s))
}

type mockIterator struct {
	NextFunc  func(interface{}) error
	CloseFunc func() error
}

var _ iterator = &mockIterator{}

func (i *mockIterator) Next(ifce interface{}) error {
	return i.NextFunc(ifce)
}

func (i *mockIterator) Close() error {
	return i.CloseFunc()
}

type mockChanges struct {
	NextFunc  func(*driver.Change) error
	CloseFunc func() error
}

var _ driver.Changes = &mockChanges{}

func (c *mockChanges) Next(ch *driver.Change) error {
	return c.NextFunc(ch)
}

func (c *mockChanges) Close() error {
	return c.CloseFunc()
}
