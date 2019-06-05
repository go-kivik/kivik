// +build !1.13

package kivik

import (
	"golang.org/x/xerrors"
)

type printer = xerrors.Printer

var formatError = xerrors.FormatError
