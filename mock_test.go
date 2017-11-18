package kivik

import (
	"context"
	"errors"
	"io"

	"github.com/flimzy/kivik/driver"
)

type mockDB struct {
	driver.DB
}

var _ driver.DB = &mockDB{}

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
