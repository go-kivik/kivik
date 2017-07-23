package memory

import (
	"context"
	"testing"
)

func TestCreateIndex(t *testing.T) {
	d := &db{}
	err := d.CreateIndex(context.Background(), "foo", "bar", "baz")
	if err != errFindNotImplemented {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestGetIndexes(t *testing.T) {
	d := &db{}
	_, err := d.GetIndexes(context.Background())
	if err != errFindNotImplemented {
		t.Errorf("Unexpected error: %s", err)
	}
}

func TestDeleteIndex(t *testing.T) {
	d := &db{}
	err := d.DeleteIndex(context.Background(), "foo", "bar")
	if err != errFindNotImplemented {
		t.Errorf("Unexpected error: %s", err)
	}
}
