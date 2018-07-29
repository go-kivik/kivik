package kivik

type statusCoder interface {
	StatusCode() int
}

// StatusCode returns the HTTP status code embedded in the error, or 500
// (internal server error), if there was no specified status code.  If err is
// nil, StatusCode returns 0. This provides a convenient way to determine the
// precise nature of a Kivik-returned error.
//
// For example, to panic for all but NotFound errors:
//
//  row, err := db.Get(context.TODO(), "docID")
//  if kivik.StatusCode(err) == kivik.StatusNotFound {
//      return
//  }
//  if err != nil {
//      panic(err)
//  }
//
// This method uses the statusCoder interface, which is not exported by this
// package, but is considered part of the stable public API.  Driver
// implementations are expected to return errors which conform to this
// interface.
//
//  type statusCoder interface {
//      StatusCode() int
//  }
func StatusCode(err error) int {
	if err == nil {
		return 0
	}
	if coder, ok := err.(statusCoder); ok {
		return coder.StatusCode()
	}
	return StatusInternalServerError
}

type exitStatuser interface {
	ExitStatus() int
}

// ExitStatus returns the curl exit status embedded in the error, or 1 (unknown
// error), if there was no specified exit status.  If err is nil, ExitStatus
// returns 0.
func ExitStatus(err error) int {
	if err == nil {
		return 0
	}
	if statuser, ok := err.(exitStatuser); ok {
		return statuser.ExitStatus()
	}
	return ExitUnknownFailure
}

type reasoner interface {
	Reason() string
}

// Reason returns the reason description for the error, or the error itself
// if none. A nil error returns an empty string.
func Reason(err error) string {
	if err == nil {
		return ""
	}
	if r, ok := err.(reasoner); ok {
		return r.Reason()
	}
	return err.Error()
}
