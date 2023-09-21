package memorydb

type statusError struct {
	error
	status int
}

func (e statusError) Unwrap() error   { return e.error }
func (e statusError) HTTPStatus() int { return e.status }
