package driver

type Iterator interface {
	// SetValue is called once, and should return a zero-value for the iterator
	// type.
	SetValue() interface{}
	Next(interface{}) error
	Close() error
}
