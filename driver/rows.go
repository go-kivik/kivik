package driver

import "encoding/json"

// Row is a generic view result row.
type Row struct {
	// ID is the document ID of the result.
	ID string `json:"id"`
	// Key is the view key of the result. For built-in views, this is the same
	// as ID.
	Key string `json:"key"`
	// Value is the raw, un-decoded JSON value. For most built-in views (such as
	// /_all_docs), this is `{"rev":"X-xxx"}`.
	Value json.RawMessage `json:"value"`
	// Doc is the raw, un-decoded JSON document. This is only populated by views
	// which return docs, such as /_all_docs?include_docs=true.
	Doc json.RawMessage `json:"doc"`
	// Changes represents a list of document leaf revisions for the /_changes
	// endpoint.
	Changes Changes `json:"changes"`
}

// Changes represents a "changes" field of a result in the /_changes stream.
type Changes []string

// UnmarshalJSON fulfils the json.Unmarshaler interface
func (c *Changes) UnmarshalJSON(data []byte) error {
	var changes []struct {
		Rev string `json:"rev"`
	}
	if err := json.Unmarshal(data, &changes); err != nil {
		return err
	}
	revs := Changes(make([]string, len(changes)))
	for i, change := range changes {
		revs[i] = change.Rev
	}
	*c = revs
	return nil
}

// Rows is an iterator over a view's results.
type Rows interface {
	// Offset is the offset where the result set starts.
	Offset() int64
	// TotalRows is the number of documents in the database/view.
	TotalRows() int64
	// UpdateSeq is the update sequence of the database, if requested in the
	// result set.
	UpdateSeq() string
	// Next is called to populate row with the values of the next row in a
	// result set.
	//
	// Next should return io.EOF when there are no more rows.
	Next(*Row) error
	// Close closes the rows iterator.
	Close() error
}
