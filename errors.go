package kivik

type kivikError string

func (e kivikError) Error() string {
	return string(e)
}

// NotImplemented is returned as an error if the underlying driver does not
// implement an optional method.
const NotImplemented kivikError = "kivik: UUIDs() not implemented by driver"
