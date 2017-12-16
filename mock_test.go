package kivik

import (
	"errors"
	"io"
	"io/ioutil"
	"strings"

	"github.com/go-kivik/kivik/driver"
)

type errReader string

var _ io.ReadCloser = errReader("")

func (r errReader) Close() error {
	return nil
}

func (r errReader) Read(_ []byte) (int, error) {
	return 0, errors.New(string(r))
}

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

type mockDBUpdates struct {
	id        string
	NextFunc  func(*driver.DBUpdate) error
	CloseFunc func() error
}

var _ driver.DBUpdates = &mockDBUpdates{}

func (u *mockDBUpdates) Close() error {
	return u.CloseFunc()
}

func (u *mockDBUpdates) Next(dbupdate *driver.DBUpdate) error {
	return u.NextFunc(dbupdate)
}
