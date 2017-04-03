package bindings

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

type pouchError struct {
	*js.Object
	Err     string
	Message string
	Status  int
}

func NewPouchError(o *js.Object) error {
	var err, msg string
	switch {
	case o.Get("reason") != js.Undefined:
		msg = o.Get("reason").String()
	case o.Get("message") != js.Undefined:
		msg = o.Get("message").String()
	}
	switch {
	case o.Get("name") != js.Undefined:
		err = o.Get("name").String()
	case o.Get("error") != js.Undefined:
		err = o.Get("error").String()
	}
	return &pouchError{
		Err:     err,
		Message: msg,
		Status:  o.Get("status").Int(),
	}
}

func (e *pouchError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Message)
}

func (e *pouchError) StatusCode() int {
	return e.Status
}
