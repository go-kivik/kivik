package driver

import (
	"context"
	"encoding/json"
)

// SearchInfo is the result of a SearchInfo request.
type SearchInfo struct {
	Name        string
	SearchIndex SearchIndex
	// RawMessage is the raw JSON response returned by the server.
	RawResponse json.RawMessage
}

// SearchIndex contains textual search index informatoin.
type SearchIndex struct {
	PendingSeq   int64
	DocDelCount  int64
	DocCount     int64
	DiskSize     int64
	CommittedSeq int64
}

// Searcher is an optional interface, which may be satisfied by a DB to support
// full-text lucene searches, as added in CouchDB 3.0.0.
type Searcher interface {
	// Search performs a full-text search against the specified ddoc and index,
	// with the specified Lucene query.
	Search(ctx context.Context, ddoc, index, query string, options map[string]interface{}) (Rows, error)
	// SearchInfo returns statistics about the specified search index.
	SearchInfo(ctx context.Context, ddoc, index string) (*SearchInfo, error)
	// SerachAnalyze tests the results of Lucene analyzer tokenization on sample text.
	SearchAnalyze(ctx context.Context, text string) ([]string, error)
}
