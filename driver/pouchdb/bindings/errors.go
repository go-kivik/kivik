package bindings

import (
	"fmt"

	"github.com/gopherjs/gopherjs/js"
)

type pouchError struct {
	*js.Object
	Err     string `js:"error"`
	Reason  string `js:"reason"`
	Name    string `js:"name"`
	Status  int    `js:"status"`
	Message string `js:"message"`
}

func (e *pouchError) Error() string {
	return fmt.Sprintf("%s: %s", e.Err, e.Reason)
}

func (e *pouchError) StatusCode() int {
	return e.Status
}
