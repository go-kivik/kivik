package kivik

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kivik/kivik/v4/driver"
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

var searchNotImplemented = &Error{HTTPStatus: http.StatusNotImplemented, Message: "kivik: driver does not support Search interface"}

// Search performs a full-text Lucene search against the named index in the
// specified design document.
//
// See https://docs.couchdb.org/en/stable/api/ddoc/search.html#get--db-_design-ddoc-_search-index
func (db *DB) Search(ctx context.Context, ddoc, index, query string, options ...Options) (*Rows, error) {
	if searcher, ok := db.driverDB.(driver.Searcher); ok {
		rowsi, err := searcher.Search(ctx, ddoc, index, query, mergeOptions(options...))
		if err != nil {
			return nil, err
		}
		return newRows(ctx, rowsi), nil
	}
	return nil, searchNotImplemented
}

// SearchInfo returns statistics about the named index in the specified design
// document.
//
// See https://docs.couchdb.org/en/stable/api/ddoc/search.html#get--db-_design-ddoc-_search_info-index
func (db *DB) SearchInfo(ctx context.Context, ddoc, index string) (*SearchInfo, error) {
	if searcher, ok := db.driverDB.(driver.Searcher); ok {
		info, err := searcher.SearchInfo(ctx, ddoc, index)
		if err != nil {
			return nil, err
		}
		return &SearchInfo{
			Name:        info.Name,
			RawResponse: info.RawResponse,
			SearchIndex: SearchIndex(info.SearchIndex),
		}, nil
	}
	return nil, searchNotImplemented
}

// SearchAnalyze tests the results of Lucene analyzer tokenization on the
// provided sample text.
//
// See https://docs.couchdb.org/en/stable/api/server/common.html#post--_search_analyze
func (db *DB) SearchAnalyze(ctx context.Context, text string) ([]string, error) {
	if searcher, ok := db.driverDB.(driver.Searcher); ok {
		return searcher.SearchAnalyze(ctx, text)
	}
	return nil, searchNotImplemented
}
