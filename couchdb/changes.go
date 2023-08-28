// Licensed under the Apache License, Version 2.0 (the "License"); you may not
// use this file except in compliance with the License. You may obtain a copy of
// the License at
//
//  http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations under
// the License.

package couchdb

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-kivik/couchdb/v4/chttp"
	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/driver"
)

// Changes returns the changes stream for the database.
func (d *db) Changes(ctx context.Context, opts map[string]interface{}) (driver.Changes, error) {
	key := "results"
	if f, ok := opts["feed"]; ok {
		if f == "eventsource" {
			return nil, &kivik.Error{Status: http.StatusBadRequest, Err: errors.New("kivik: eventsource feed not supported, use 'continuous'")}
		}
		if f == "continuous" {
			key = ""
		}
	}
	query, err := optionsToParams(opts)
	if err != nil {
		return nil, err
	}
	options := &chttp.Options{
		Query: query,
	}
	resp, err := d.Client.DoReq(ctx, http.MethodGet, d.path("_changes"), options)
	if err != nil {
		return nil, err
	}
	if err = chttp.ResponseError(resp); err != nil {
		return nil, err
	}
	etag, _ := chttp.ETag(resp)
	return newChangesRows(ctx, key, resp.Body, etag), nil
}

type continuousChangesParser struct{}

func (p *continuousChangesParser) parseMeta(i interface{}, dec *json.Decoder, key string) error {
	meta := i.(*changesMeta)
	return meta.parseMeta(key, dec)
}

func (p *continuousChangesParser) decodeItem(i interface{}, dec *json.Decoder) error {
	row := i.(*driver.Change)
	ch := &change{Change: row}
	if err := dec.Decode(ch); err != nil {
		return &kivik.Error{Status: http.StatusBadGateway, Err: err}
	}
	ch.Change.Seq = string(ch.Seq)
	return nil
}

type changesMeta struct {
	lastSeq sequenceID
	pending int64
}

// parseMeta parses result metadata
func (m *changesMeta) parseMeta(key string, dec *json.Decoder) error {
	switch key {
	case "last_seq":
		return dec.Decode(&m.lastSeq)
	case "pending":
		return dec.Decode(&m.pending)
	default:
		// Just consume the value, since we don't know what it means.
		var discard json.RawMessage
		return dec.Decode(&discard)
	}
}

type changesRows struct {
	*iter
	etag string
}

func newChangesRows(ctx context.Context, key string, r io.ReadCloser, etag string) *changesRows {
	var meta *changesMeta
	if key != "" {
		meta = &changesMeta{}
	}
	return &changesRows{
		iter: newIter(ctx, meta, key, r, &continuousChangesParser{}),
		etag: etag,
	}
}

var _ driver.Changes = &changesRows{}

type change struct {
	*driver.Change
	Seq sequenceID `json:"seq"`
}

func (r *changesRows) Next(row *driver.Change) error {
	row.Deleted = false
	return r.iter.next(row)
}

// LastSeq returns the last sequence ID.
func (r *changesRows) LastSeq() string {
	return string(r.iter.meta.(*changesMeta).lastSeq)
}

// Pending returns the pending count.
func (r *changesRows) Pending() int64 {
	return r.iter.meta.(*changesMeta).pending
}

// ETag returns the unquoted ETag header for the CouchDB response, if any.
func (r *changesRows) ETag() string {
	return r.etag
}
