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
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	kivik "github.com/go-kivik/kivik/v4"
	"github.com/go-kivik/kivik/v4/couchdb/chttp"
	"github.com/go-kivik/kivik/v4/driver"
	internal "github.com/go-kivik/kivik/v4/int/errors"
)

func (c *client) AllDBs(ctx context.Context, opts driver.Options) ([]string, error) {
	query := url.Values{}
	opts.Apply(&query)
	var allDBs []string
	err := c.DoJSON(ctx, http.MethodGet, "/_all_dbs", &chttp.Options{Query: query}, &allDBs)
	return allDBs, err
}

func (c *client) DBExists(ctx context.Context, dbName string, _ driver.Options) (bool, error) {
	if dbName == "" {
		return false, missingArg("dbName")
	}
	_, err := c.DoError(ctx, http.MethodHead, url.PathEscape(dbName), nil)
	if kivik.HTTPStatus(err) == http.StatusNotFound {
		return false, nil
	}
	return err == nil, err
}

func (c *client) CreateDB(ctx context.Context, dbName string, opts driver.Options) error {
	if dbName == "" {
		return missingArg("dbName")
	}
	query := url.Values{}
	opts.Apply(&query)
	_, err := c.DoError(ctx, http.MethodPut, url.PathEscape(dbName), &chttp.Options{Query: query})
	return err
}

func (c *client) DestroyDB(ctx context.Context, dbName string, _ driver.Options) error {
	if dbName == "" {
		return missingArg("dbName")
	}
	_, err := c.DoError(ctx, http.MethodDelete, url.PathEscape(dbName), nil)
	return err
}

func (c *client) DBUpdates(ctx context.Context, opts driver.Options) (updates driver.DBUpdates, err error) {
	query := url.Values{}
	opts.Apply(&query)
	// TODO #864: Remove this default behavior for v5
	if _, ok := query["feed"]; !ok {
		query.Set("feed", "continuous")
	} else if query.Get("feed") == "" {
		delete(query, "feed")
	}
	if _, ok := query["since"]; !ok {
		query.Set("since", "now")
	} else if query.Get("since") == "" {
		delete(query, "since")
	}
	if query.Get("feed") == "eventsource" {
		return nil, &internal.Error{Status: http.StatusBadRequest, Message: "kivik: eventsource feed type not supported by CouchDB driver"}
	}
	resp, err := c.DoReq(ctx, http.MethodGet, "/_db_updates", &chttp.Options{Query: query})
	if err != nil {
		return nil, err
	}
	if err := chttp.ResponseError(resp); err != nil {
		return nil, err
	}
	return newUpdates(ctx, resp.Body)
}

type couchUpdates struct {
	*iter
	meta *updatesMeta
}

var _ driver.LastSeqer = &couchUpdates{}

type updatesMeta struct {
	lastSeq string
}

func (u *updatesMeta) parseMeta(key interface{}, dec *json.Decoder) error {
	if key == "last_seq" {
		return dec.Decode(&u.lastSeq)
	}
	// Just consume the value, since we don't know what it means.
	var discard json.RawMessage
	return dec.Decode(&discard)
}

type updatesParser struct{}

var _ parser = &updatesParser{}

func (p *updatesParser) decodeItem(i interface{}, dec *json.Decoder) error {
	return dec.Decode(i)
}

func (p *updatesParser) parseMeta(i interface{}, dec *json.Decoder, key string) error {
	meta := i.(*updatesMeta)
	return meta.parseMeta(key, dec)
}

func newUpdates(ctx context.Context, r io.ReadCloser) (driver.DBUpdates, error) {
	r, feedType, err := updatesFeedType(r)
	if err != nil {
		return nil, &internal.Error{Status: http.StatusBadGateway, Err: err}
	}

	switch feedType {
	case feedTypeContinuous:
		return newContinuousUpdates(ctx, r), nil
	case feedTypeNormal:
		return newNormalUpdates(ctx, r), nil
	case feedTypeEmpty:
		return newEmptyUpdates(r)
	}
	panic("unknown") // TODO: test
}

type feedType int

const (
	feedTypeNormal feedType = iota
	feedTypeContinuous
	feedTypeEmpty
)

// updatesFeedType detects the type of Updates Feed (continuous, or normal) by
// reading the first two JSON tokens of the stream. It then returns a
// re-assembled reader with the results.
func updatesFeedType(r io.ReadCloser) (io.ReadCloser, feedType, error) {
	buf := &bytes.Buffer{}
	// Read just the first two tokens, to determine feed type. We use a
	// json.Decoder rather than reading raw bytes, because there may be
	// whitespace, and it's also nice to return a consistent error in case the
	// body is invalid.
	dec := json.NewDecoder(io.TeeReader(r, buf))
	tok, err := dec.Token()
	if err != nil {
		return nil, 0, err
	}
	if tok != json.Delim('{') {
		return nil, 0, fmt.Errorf("kivik: unexpected JSON token %q in feed response, expected `{`", tok)
	}
	tok, err = dec.Token()
	if err != nil {
		return nil, 0, err
	}

	// restore reader
	r = struct {
		io.Reader
		io.Closer
	}{
		Reader: io.MultiReader(buf, r),
		Closer: r,
	}

	switch tok {
	case "db_name": // Continuous feed
		return r, feedTypeContinuous, nil
	case "results": // Normal feed
		return r, feedTypeNormal, nil
	case "last_seq": // Normal feed, but with no results
		return r, feedTypeEmpty, nil
	default: // Something unexpected
		return nil, 0, fmt.Errorf("kivik: unexpected JSON token %q in feed response, expected `db_name` or `results`", tok)
	}
}

func newNormalUpdates(ctx context.Context, r io.ReadCloser) *couchUpdates {
	meta := &updatesMeta{}
	return &couchUpdates{
		iter: newIter(ctx, meta, "results", r, &updatesParser{}),
		meta: meta,
	}
}

func newContinuousUpdates(ctx context.Context, r io.ReadCloser) *couchUpdates {
	return &couchUpdates{
		iter: newIter(ctx, nil, "", r, &updatesParser{}),
	}
}

func newEmptyUpdates(r io.ReadCloser) (driver.DBUpdates, error) {
	var updates emptyUpdates
	if err := json.NewDecoder(r).Decode(&updates); err != nil {
		return nil, err
	}
	return &updates, nil
}

type emptyUpdates struct {
	Lastseq string `json:"last_seq"`
}

var (
	_ driver.DBUpdates = (*emptyUpdates)(nil)
	_ driver.LastSeqer = (*emptyUpdates)(nil)
)

func (u *emptyUpdates) Close() error                { return nil }
func (u *emptyUpdates) Next(*driver.DBUpdate) error { return io.EOF }
func (u *emptyUpdates) LastSeq() (string, error)    { return u.Lastseq, nil }

func (u *couchUpdates) Next(update *driver.DBUpdate) error {
	return u.iter.next(update)
}

func (u *couchUpdates) LastSeq() (string, error) {
	if u.meta == nil {
		return "", nil
	}
	return u.meta.lastSeq, nil
}

// Ping queries the /_up endpoint, and returns true if there are no errors, or
// if a 400 (Bad Request) is returned, and the Server: header indicates a server
// version prior to 2.x.
func (c *client) Ping(ctx context.Context) (bool, error) {
	resp, err := c.DoError(ctx, http.MethodHead, "/_up", nil)
	if kivik.HTTPStatus(err) == http.StatusBadRequest {
		return strings.HasPrefix(resp.Header.Get("Server"), "CouchDB/1."), nil
	}
	if kivik.HTTPStatus(err) == http.StatusNotFound {
		return false, nil
	}
	return err == nil, err
}
