package driver

import "encoding/json"

// Row is a generic view result row.
type Row struct {
	ID    string          `json:"id"`
	Key   string          `json:"key"`
	Value json.RawMessage `json:"value"`
}

// Rows is an iterator over a view's results.
type Rows interface {
	// Offset is the offset where the result set starts.
	Offset() int64
	// TotalRows is the number of documents in the database/view.
	TotalRows() int64
	// Next is called to populate row with the values of the next row in a
	// result set.
	//
	// Next should return io.EOF when there are no more rows.
	Next(*Row) error
	// Close closes the rows iterator.
	Close() error
}
