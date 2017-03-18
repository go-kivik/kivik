package couchdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/flimzy/kivik"
	"github.com/flimzy/kivik/driver"
	"github.com/flimzy/kivik/driver/couchdb/chttp"
	"github.com/flimzy/kivik/driver/ouchdb"
)

type db struct {
	*client
	dbName      string
	forceCommit bool
}

func (d *db) path(path string, query url.Values) string {
	url, _ := url.Parse(d.dbName + "/" + strings.TrimPrefix(path, "/"))
	if query != nil {
		url.RawQuery = query.Encode()
	}
	return url.String()
}

func optionsToParams(opts map[string]interface{}) (url.Values, error) {
	params := url.Values{}
	for key, i := range opts {
		var values []string
		switch i.(type) {
		case string:
			values = []string{i.(string)}
		case []string:
			values = i.([]string)
		default:
			return nil, fmt.Errorf("Cannot convert type %T to []string", i)
		}
		for _, value := range values {
			params.Add(key, value)
		}
	}
	return params, nil
}

// AllDocsContext returns all of the documents in the database.
func (d *db) AllDocsContext(ctx context.Context, docs interface{}, opts map[string]interface{}) (offset, totalrows int, seq string, err error) {
	resp, err := d.Client.DoReq(ctx, kivik.MethodGet, d.path("_all_docs", nil), nil)
	if err != nil {
		return 0, 0, "", err
	}
	if err = chttp.ResponseError(resp.Response); err != nil {
		return 0, 0, "", err
	}
	defer resp.Body.Close()
	return ouchdb.AllDocs(resp.Body, docs)
}

// GetContext fetches the requested document.
func (d *db) GetContext(ctx context.Context, docID string, doc interface{}, opts map[string]interface{}) error {
	params, err := optionsToParams(opts)
	if err != nil {
		return err
	}
	_, err = d.Client.DoJSON(ctx, http.MethodGet, d.path(docID, params), &chttp.Options{Accept: "application/json; multipart/mixed"}, doc)
	return err
}

func (d *db) CreateDocContext(ctx context.Context, doc interface{}) (docID, rev string, err error) {
	result := struct {
		ID  string `json:"id"`
		Rev string `json:"rev"`
	}{}
	opts := &chttp.Options{
		JSON:        doc,
		ForceCommit: d.forceCommit,
	}
	_, err = d.Client.DoJSON(ctx, kivik.MethodPost, d.dbName, opts, &result)
	return result.ID, result.Rev, err
}

func (d *db) PutContext(ctx context.Context, docID string, doc interface{}) (rev string, err error) {
	result := struct {
		Rev string `json:"rev"`
	}{}
	opts := &chttp.Options{
		JSON:        doc,
		ForceCommit: d.forceCommit,
	}
	_, err = d.Client.DoJSON(ctx, kivik.MethodPut, d.path(docID, nil), opts, &result)
	return result.Rev, err
}

func (d *db) DeleteContext(ctx context.Context, docID, rev string) (string, error) {
	query := url.Values{}
	query.Add("rev", rev)
	var result struct {
		Rev string `json:"rev"`
	}
	opts := &chttp.Options{
		ForceCommit: d.forceCommit,
	}
	_, err := d.Client.DoJSON(ctx, kivik.MethodDelete, d.path(docID, query), opts, &result)
	return result.Rev, err
}

func (d *db) FlushContext(ctx context.Context) (time.Time, error) {
	result := struct {
		T int64 `json:"instance_start_time,string"`
	}{}
	_, err := d.Client.DoJSON(ctx, kivik.MethodPost, d.path("/_ensure_full_commit", nil), nil, &result)
	return time.Unix(0, 0).Add(time.Duration(result.T) * time.Microsecond), err
}

func (d *db) InfoContext(ctx context.Context) (*driver.DBInfo, error) {
	result := struct {
		driver.DBInfo
		Sizes struct {
			File     int64 `json:"file"`
			External int64 `json:"external"`
			Active   int64 `json:"active"`
		} `json:"sizes"`
		UpdateSeq json.RawMessage `json:"update_seq"`
	}{}
	_, err := d.Client.DoJSON(ctx, kivik.MethodGet, d.dbName, nil, &result)
	info := result.DBInfo
	if result.Sizes.File > 0 {
		info.DiskSize = result.Sizes.File
	}
	if result.Sizes.External > 0 {
		info.ExternalSize = result.Sizes.External
	}
	if result.Sizes.Active > 0 {
		info.ActiveSize = result.Sizes.Active
	}
	info.UpdateSeq = string(bytes.Trim(result.UpdateSeq, `"`))
	return &info, err
}

func (d *db) CompactContext(ctx context.Context) error {
	res, err := d.Client.DoReq(ctx, kivik.MethodPost, d.path("/_compact", nil), nil)
	if err != nil {
		return err
	}
	return chttp.ResponseError(res.Response)
}

func (d *db) CompactViewContext(ctx context.Context, ddocID string) error {
	res, err := d.Client.DoReq(ctx, kivik.MethodPost, d.path("/_compact/"+ddocID, nil), nil)
	if err != nil {
		return err
	}
	return chttp.ResponseError(res.Response)
}

func (d *db) ViewCleanupContext(ctx context.Context) error {
	res, err := d.Client.DoReq(ctx, kivik.MethodPost, d.path("/_view_cleanup", nil), nil)
	if err != nil {
		return err
	}
	return chttp.ResponseError(res.Response)
}

func (d *db) SecurityContext(ctx context.Context) (*driver.Security, error) {
	var sec *driver.Security
	_, err := d.Client.DoJSON(ctx, kivik.MethodGet, d.path("/_security", nil), nil, &sec)
	return sec, err
}

func (d *db) SetSecurityContext(ctx context.Context, security *driver.Security) error {
	res, err := d.Client.DoReq(ctx, kivik.MethodPut, d.path("/_security", nil), &chttp.Options{JSON: security})
	if err != nil {
		return err
	}
	defer res.Body.Close()
	return chttp.ResponseError(res.Response)
}

// RevContext returns the most current rev of the requested document.
func (d *db) RevContext(ctx context.Context, docID string) (rev string, err error) {
	res, err := d.Client.DoError(ctx, http.MethodHead, d.path(docID, nil), nil)
	if err != nil {
		return "", err
	}
	return strings.Trim(res.Header.Get("Etag"), `""`), nil
}

// RevsLimitContext returns the maximum number of document revisions that will
// be tracked.
func (d *db) RevsLimitContext(ctx context.Context) (limit int, err error) {
	_, err = d.Client.DoJSON(ctx, http.MethodGet, d.path("/_revs_limit", nil), nil, &limit)
	return limit, err
}

func (d *db) SetRevsLimitContext(ctx context.Context, limit int) error {
	body, err := json.Marshal(limit)
	if err != nil {
		return err
	}
	_, err = d.Client.DoError(ctx, http.MethodPut, d.path("/_revs_limit", nil), &chttp.Options{Body: bytes.NewBuffer(body)})
	return err
}
